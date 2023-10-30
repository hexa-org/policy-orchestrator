package orchestrator

import (
	"encoding/json"
	"errors"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	log "golang.org/x/exp/slog"
	"net/http"
	"strings"
)

type ApplicationsService struct {
	ApplicationsGateway ApplicationsDataGateway
	IntegrationsGateway IntegrationsDataGateway
	Providers           map[string]Provider
}

type rarMapper struct {
	// This assumes the primary key is Resource
	// Sort key is Action
	// {
	//  "pk": {"attrName": "Resource", "attrValue": 1  }, // attrValue can only be a scalar string, or int
	//  "sk": {"attrName": "Resource", "attrValue": 1  }, // attrValue can only be a scalar string, or int
	//	"resource": {"attrName": "Policy.Nested.Resource", "attrValue": 1  }	// string or int
	//  "actions": {"attrName": "Policy.Nested.Action", "attrValue": "example"  } // string or int or array of string or int
	//  "members": {"attrName": "Policy.Nested.Members", "attrValue": "["mem1", "mem2]"  } // string or int or array of string or int
	//}

	// What if Resource, Action are not pk, sk?
	// Scan should not care about the pk, sk
	// Update item needs to pk, sk
	//attributeAccessors map[string]dynamodbAttributeAccessors
	tableDefinition dynamodbpolicystore.TableDefinition
	shape           interface{}
}

func (rm rarMapper) MapTo() (rar.ResourceActionRoles, error) {
	shape, ok := rm.shape.(map[string]interface{})
	if !ok {
		return rar.ResourceActionRoles{}, errors.New("failed to convert db item shape interface")
	}
	resource, ok := shape[rm.tableDefinition.ResourceAttrName].(string)
	if !ok {
		return rar.ResourceActionRoles{}, errors.New("failed to convert db item shape.resource to string")
	}

	actions, ok := shape[rm.tableDefinition.ActionAttrName].([]string)
	if !ok {
		return rar.ResourceActionRoles{}, errors.New("failed to convert db item shape.actions to []string")
	}
	members, ok := shape[rm.tableDefinition.MembersAttrName].([]string)
	if !ok {
		return rar.ResourceActionRoles{}, errors.New("failed to convert db item shape.members to []string")
	}

	return rar.NewResourceActionRoles(resource, actions, members)
}

func (service ApplicationsService) GatherRecords(identifier string) (ApplicationInfo, IntegrationInfo, Provider, error) {
	applicationRecord, err := service.ApplicationsGateway.FindById(identifier)
	if err != nil {
		return ApplicationInfo{}, IntegrationInfo{}, nil, err
	}
	application := ApplicationInfo{ObjectID: applicationRecord.ObjectId, Name: applicationRecord.Name, Description: applicationRecord.Description, Service: applicationRecord.Service}

	integrationRecord, err := service.IntegrationsGateway.FindById(applicationRecord.IntegrationId)
	if err != nil {
		return ApplicationInfo{}, IntegrationInfo{}, nil, err
	}
	integration := IntegrationInfo{Name: integrationRecord.Name, Key: integrationRecord.Key}

	// SAURABH - temporary workaround until we implement an app onboarding flow
	var aProvider Provider
	if integrationRecord.Provider == "amazon" {
		// One must be pk, but only one
		// If sk defined, can have multiple.
		// attributes can be of valType string, int, []string, []int
		// "attributes" must have 3 keys i.e. "resource", "actions", "members"
		itemJson := `
			{
				"metadata": {
					"pk": { "attribute": "resource" },
					"sk": { "attribute": ["actions", "another1", "another2"] }
				},
				"attributes": {
					"resource": { "nameOrPath": "Policy/Nested/ResourceX", "valType": "string" },
					"actions": { "nameOrPath": "Policy/Nested/ActionsX", "valType": "string[]" },
					"members": { "nameOrPath": "Policy/Nested/Members", "valType": "int[]" }
				}  
			}`

		//itemJson := `{ "Resource": "A-Resource", "Action": "An-Action", "Members": "some member", "Nested": { "Resource": "Child-Resource" }	}`
		decoder := json.NewDecoder(strings.NewReader(itemJson))
		decoder.UseNumber()
		var shape interface{}
		_ = decoder.Decode(&shape)
		aProvider, _ = NewOrchestrationProvider(integration.Key, integration.Key)
		// cannot import awsapigw due to import cycle
		//		aProvider, err = awsapigw.NewAwsApiGatewayProviderV2(integration.Key, integration.Key)
	} else {
		aProvider = service.Providers[strings.ToLower(integrationRecord.Provider)]
	}

	if err != nil {
		log.Error("GatherRecords", "msg", "error creating Provdider", "provider", integrationRecord.Provider, "error", err)
		return ApplicationInfo{}, IntegrationInfo{}, nil, err
	}
	return application, integration, aProvider, err // todo - test for lower?
}

func (service ApplicationsService) Apply(jsonRequest Orchestration) error {
	fromApplication, fromIntegration, fromProvider, fromErr := service.GatherRecords(jsonRequest.From)
	if fromErr != nil {
		return fromErr
	}

	toApplication, toIntegration, toProvider, toErr := service.GatherRecords(jsonRequest.To)
	if toErr != nil {
		return toErr
	}

	if !orchestrationSupported(toProvider, fromProvider) { // note - this should be temporary
		return errors.New("orchestration across providers is a work in progress and currently supports azure and google cloud")
	}

	fromPolicies, getFroErr := fromProvider.GetPolicyInfo(fromIntegration, fromApplication)
	if getFroErr != nil {
		return getFroErr
	}

	if isBetweenAmazonAndAzure(toProvider, fromProvider) {
		status, setErr := toProvider.SetPolicyInfo(toIntegration, toApplication, fromPolicies)
		if setErr != nil || status != http.StatusCreated {
			return setErr
		}
		return nil
	}

	toPolicies, getToErr := toProvider.GetPolicyInfo(toIntegration, toApplication)
	if getToErr != nil {
		return getToErr
	}

	if onlyWorksWithGoogleAndAzure(toProvider, fromProvider) {
		if !verifyAllMembersAreUsers(fromPolicies) { // note - this should be temporary
			return errors.New("orchestration across providers with domain members is a work in progress")
		}
	}

	modifiedPolicies, err := service.RetainResource(fromPolicies, toPolicies)
	if err != nil {
		return err
	}

	if onlyWorksWithGoogleAndAzure(toProvider, fromProvider) {
		modifiedPolicies, err = service.RetainAction(modifiedPolicies, toPolicies)
		if err != nil {
			return err
		}
	}

	status, setErr := toProvider.SetPolicyInfo(toIntegration, toApplication, modifiedPolicies)
	if setErr != nil || status != http.StatusCreated {
		return setErr
	}
	return nil
}

func (service ApplicationsService) RetainResource(fromPolicies, toPolicies []hexapolicy.PolicyInfo) ([]hexapolicy.PolicyInfo, error) {
	var firstResourceId string

	resourceIds := make([]string, 0)
	for _, policy := range toPolicies {
		if firstResourceId == "" {
			firstResourceId = policy.Object.ResourceID
		}
		resourceIds = append(resourceIds, policy.Object.ResourceID)
	}

	for _, foundResourceId := range resourceIds {
		if firstResourceId != foundResourceId {
			return []hexapolicy.PolicyInfo{}, errors.New("sorry, found more than one resource id within policies")
		}
	}

	modified := make([]hexapolicy.PolicyInfo, 0)
	for _, policy := range fromPolicies {
		policy.Object.ResourceID = firstResourceId
		modified = append(modified, policy)
	}
	return modified, nil
}

func (service ApplicationsService) RetainAction(fromPolicies, toPolicies []hexapolicy.PolicyInfo) ([]hexapolicy.PolicyInfo, error) {
	firstActionUri := toPolicies[0].Actions[0].ActionUri // todo update to handle all action uris from toProvider

	modified := make([]hexapolicy.PolicyInfo, 0)
	for _, policy := range fromPolicies {
		policy.Actions = make([]hexapolicy.ActionInfo, 1)
		policy.Actions[0].ActionUri = firstActionUri
		modified = append(modified, policy)
	}
	return modified, nil
}

func verifyAllMembersAreUsers(policies []hexapolicy.PolicyInfo) bool {
	var areMembersUsers bool
	for _, policy := range policies {
		for _, member := range policy.Subject.Members {
			areMembersUsers = strings.Contains(member, "user:")
		}
	}
	return areMembersUsers
}

func orchestrationSupported(toProvider Provider, fromProvider Provider) bool {
	if toProvider == fromProvider {
		return true
	}
	return onlyWorksWithGoogleAndAzure(toProvider, fromProvider)
}

func onlyWorksWithGoogleAndAzure(toProvider Provider, fromProvider Provider) bool {
	if toProvider.Name() == "google_cloud" && fromProvider.Name() == "azure" {
		return true
	}
	if toProvider.Name() == "azure" && fromProvider.Name() == "google_cloud" {
		return true
	}
	// TODO - Check this with Gerry
	if isBetweenAmazonAndAzure(toProvider, fromProvider) {
		return true
	}
	return false
}

func isBetweenAmazonAndAzure(toProvider Provider, fromProvider Provider) bool {
	if toProvider.Name() == "amazon" && fromProvider.Name() == "azure" {
		return true
	}
	if toProvider.Name() == "azure" && fromProvider.Name() == "amazon" {
		return true
	}
	return false
}
