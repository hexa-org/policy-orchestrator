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

	if toProvider != fromProvider {
		return errors.New("sorry, orchestration across providers is work in progress")
	}

	fromPolicies, getFroErr := fromProvider.GetPolicyInfo(fromIntegration, fromApplication)
	if getFroErr != nil {
		return getFroErr
	}

	toPolicies, getToErr := fromProvider.GetPolicyInfo(fromIntegration, fromApplication)
	if getToErr != nil {
		return getToErr
	}

	modifiedToPoliciesRetainingResourceId, err := service.RetainResource(fromPolicies, toPolicies)
	if err != nil {
		return err
	}

	status, setErr := toProvider.SetPolicyInfo(toIntegration, toApplication, modifiedToPoliciesRetainingResourceId)
	if setErr != nil || status != http.StatusCreated {
		return setErr
	}
	return nil
}

func (service ApplicationsService) RetainResource(fromPolicies, toPolicies []policysupport.PolicyInfo) ([]policysupport.PolicyInfo, error) {
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
		modified = append(modified, policy)
	}
	return modified, nil
}
