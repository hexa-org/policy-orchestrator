package orchestrator

import (
	"errors"
	"net/http"
	"strings"
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

	fromPolicies, getErr := fromProvider.GetPolicyInfo(fromIntegration, fromApplication)
	if getErr != nil {
		return getErr
	}

	status, setErr := toProvider.SetPolicyInfo(toIntegration, toApplication, fromPolicies)
	if setErr != nil || status != http.StatusCreated {
		return setErr
	}
	return nil
}
