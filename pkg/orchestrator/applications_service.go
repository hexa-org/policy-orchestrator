package orchestrator

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
)

type ApplicationsService struct {
	ApplicationsGateway ApplicationsDataGateway
	IntegrationsGateway IntegrationsDataGateway
	Providers           map[string]Provider
}

func (service ApplicationsService) GatherRecords(identifier string) (ApplicationInfo, IntegrationInfo, Provider, error) {
	applicationRecord, err := service.ApplicationsGateway.FindById(identifier)
	if err != nil {
		return ApplicationInfo{}, IntegrationInfo{}, nil, err
	}
	application := ApplicationInfo{ObjectID: applicationRecord.ObjectId, Name: applicationRecord.Name, Description: applicationRecord.Description}

	integrationRecord, err := service.IntegrationsGateway.FindById(applicationRecord.IntegrationId)
	if err != nil {
		return ApplicationInfo{}, IntegrationInfo{}, nil, err
	}
	integration := IntegrationInfo{Name: integrationRecord.Name, Key: integrationRecord.Key}

	return application, integration, service.Providers[strings.ToLower(integrationRecord.Provider)], err // todo - test for lower?
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

	toPolicies, getToErr := toProvider.GetPolicyInfo(toIntegration, toApplication)
	if getToErr != nil {
		return getToErr
	}

	if !verifyAllMembersAreUsers(fromPolicies) || !verifyAllMembersAreUsers(toPolicies) { // note - this should be temporary
		return errors.New("orchestration across providers with domain members is a work in progress")
	}

	modifiedToPoliciesRetainingResourceIdAndActions, err := service.RetainResourceAndAction(fromPolicies, toPolicies)
	if err != nil {
		return err
	}

	status, setErr := toProvider.SetPolicyInfo(toIntegration, toApplication, modifiedToPoliciesRetainingResourceIdAndActions)
	if setErr != nil || status != http.StatusCreated {
		return setErr
	}
	return nil
}

func (service ApplicationsService) RetainResourceAndAction(fromPolicies, toPolicies []policysupport.PolicyInfo) ([]policysupport.PolicyInfo, error) {
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
			return []policysupport.PolicyInfo{}, errors.New("sorry, found more than one resource id within policies")
		}
	}

	modified := make([]policysupport.PolicyInfo, 0)
	for _, policy := range fromPolicies {
		policy.Object.ResourceID = firstResourceId
		policy.Actions = toPolicies[0].Actions
		modified = append(modified, policy)
	}
	return modified, nil
}

func (service ApplicationsService) RetainResourceActions(fromPolicies, toPolicies []policysupport.PolicyInfo) ([]policysupport.PolicyInfo, error) {
	var resourceActions []policysupport.ActionInfo

	actions := make([]policysupport.ActionInfo, 0)
	for _, policy := range toPolicies {
		if len(resourceActions) == 0 {
			resourceActions = policy.Actions
		}
		actions = append(actions, policy.Actions...)
	}

	modified := make([]policysupport.PolicyInfo, 0)
	for _, policy := range fromPolicies {
		policy.Actions = actions
		modified = append(modified, policy)
	}
	return modified, nil
}

func verifyAllMembersAreUsers(policies []policysupport.PolicyInfo) bool {
	var areMembersUsers bool
	for _, policy := range policies {
		for _, member := range policy.Subject.Members {
			areMembersUsers = strings.Contains(member, "user:")
		}
	}
	return areMembersUsers
}

func orchestrationSupported(toProvider Provider, fromProvider Provider) bool {
	if toProvider.Name() == "google_cloud" && fromProvider.Name() == "azure" {
		return true
	}
	if toProvider.Name() == "azure" && fromProvider.Name() == "google_cloud" {
		return true
	}
	if toProvider == fromProvider {
		return true
	}
	return false
}
