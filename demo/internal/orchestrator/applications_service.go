package orchestrator

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2"
	logger "golang.org/x/exp/slog"
)

type ApplicationsService struct {
	ApplicationsGateway ApplicationsDataGateway
	IntegrationsGateway IntegrationsDataGateway
	ProviderBuilder     *providerBuilder
	DisableChecks       bool // Only set to true by tests
}

func (service ApplicationsService) GatherRecords(identifier string) (policyprovider.ApplicationInfo, policyprovider.IntegrationInfo, policyprovider.Provider, error) {
	applicationRecord, err := service.ApplicationsGateway.FindById(identifier)
	if err != nil {
		return policyprovider.ApplicationInfo{}, policyprovider.IntegrationInfo{}, nil, err
	}
	application := policyprovider.ApplicationInfo{ObjectID: applicationRecord.ObjectId, Name: applicationRecord.Name, Description: applicationRecord.Description, Service: applicationRecord.Service}

	integrationRecord, err := service.IntegrationsGateway.FindById(applicationRecord.IntegrationId)
	if err != nil {
		return policyprovider.ApplicationInfo{}, policyprovider.IntegrationInfo{}, nil, err
	}
	integration := policyprovider.IntegrationInfo{Name: integrationRecord.Name, Key: integrationRecord.Key}

	// SAURABH - temporary workaround until we implement an app onboarding flow
	var aProvider policyprovider.Provider
	if integrationRecord.Provider == "amazon" {
		resAttrDef := providersV2.NewAttributeDefinition("Resource", "string", true, false)
		actionsAttrDef := providersV2.NewAttributeDefinition("Action", "string", false, true)
		membersDef := providersV2.NewAttributeDefinition("Members", "string", false, false)
		tableDef := providersV2.NewTableDefinition(resAttrDef, actionsAttrDef, membersDef)
		policyStore := providersV2.NewDynamicItemStore(providersV2.AwsPolicyStoreTableName, integration.Key, tableDef)

		// idp := providersV2.GetAppsProvider(integrationRecord.Provider, integration.Key)
		idp := providersV2.NewCognitoIdp(integration.Key)
		aProvider, err = NewOrchestrationProvider(integrationRecord.Provider, idp, policyStore)
		if err != nil {
			logger.Error("GatherRecords", "msg", "error creating Provider", "provider", integrationRecord.Provider, "error", err)
			return policyprovider.ApplicationInfo{}, policyprovider.IntegrationInfo{}, nil, err
		}
	} else if integrationRecord.Provider == "azure" {
		idp := providersV2.NewApimAppProvider(integration.Key)
		policyStoreProvider := providersV2.NewApimPolicyProvider(integration.Key, application.ObjectID, application.Name)
		aProvider, err = NewOrchestrationProvider(integrationRecord.Provider, idp, policyStoreProvider)

		if err != nil {
			logger.Error("GatherRecords", "msg", "error creating Provider", "provider", integrationRecord.Provider, "error", err)
			return policyprovider.ApplicationInfo{}, policyprovider.IntegrationInfo{}, nil, err
		}

	} else {
		aProvider, err = service.ProviderBuilder.GetAppsProvider(strings.ToLower(integrationRecord.Provider), nil)
		if err != nil {
			logger.Error("GatherRecords", "msg", "error creating Provider", "provider", integrationRecord.Provider, "error", err)
			return policyprovider.ApplicationInfo{}, policyprovider.IntegrationInfo{}, nil, err
		}
	}

	return application, integration, aProvider, nil // todo - test for lower?
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

	logger.Info("Apply", "fromProvider", fromProvider.Name(), "toProvider", toProvider.Name(), "fromApp", fromApplication.Name, "toApp", toApplication.Name)

	if !service.DisableChecks && !orchestrationSupported(toProvider, fromProvider) { // note - this should be temporary
		return errors.New("orchestration across providers is a work in progress and currently supports azure and google cloud")
	}

	fromPolicies, getFroErr := fromProvider.GetPolicyInfo(fromIntegration, fromApplication)
	if getFroErr != nil {
		return getFroErr
	}

	if service.DisableChecks || isBetweenAmazonAndAzure(toProvider, fromProvider) {
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

	if service.DisableChecks || onlyWorksWithGoogleAndAzure(toProvider, fromProvider) {
		if !verifyAllMembersAreUsers(fromPolicies) { // note - this should be temporary
			return errors.New("orchestration across providers with domain members is a work in progress")
		}
	}

	modifiedPolicies, err := service.RetainResource(fromPolicies, toPolicies)
	if err != nil {
		return err
	}

	if service.DisableChecks || onlyWorksWithGoogleAndAzure(toProvider, fromProvider) {
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

func orchestrationSupported(toProvider policyprovider.Provider, fromProvider policyprovider.Provider) bool {
	if toProvider == fromProvider {
		return true
	}
	return onlyWorksWithGoogleAndAzure(toProvider, fromProvider)
}

func onlyWorksWithGoogleAndAzure(toProvider policyprovider.Provider, fromProvider policyprovider.Provider) bool {
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

func isBetweenAmazonAndAzure(toProvider policyprovider.Provider, fromProvider policyprovider.Provider) bool {
	if toProvider.Name() == "amazon" && fromProvider.Name() == "azure" {
		return true
	}
	if toProvider.Name() == "azure" && fromProvider.Name() == "amazon" {
		return true
	}
	return false
}
