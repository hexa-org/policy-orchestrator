package azarm

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
	log "golang.org/x/exp/slog"
	"net/http"
)

type ApimProviderService struct {
	azureClient       azad.AzureClient
	armApimSvc        azapim.ArmApimSvc
	apimNamedValueSvc apimnv.ApimNamedValueSvc
}

func NewApimProviderService(armApimSvc azapim.ArmApimSvc, azureClient azad.AzureClient, apimNamedValueSvc apimnv.ApimNamedValueSvc) *ApimProviderService {
	return &ApimProviderService{armApimSvc: armApimSvc, azureClient: azureClient, apimNamedValueSvc: apimNamedValueSvc}
}

func (s *ApimProviderService) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	azWebApps, err := s.azureClient.GetAzureApplications(info.Key)
	if err != nil {
		log.Error("ApimProviderService.DiscoverApplications", "GetAzureApplications err", err)
		return []orchestrator.ApplicationInfo{}, err
	}

	apps := make([]orchestrator.ApplicationInfo, 0)
	for _, oneApp := range azWebApps {
		log.Info("1 ApimProviderService.DiscoverApplications", "App.Name", oneApp.Name, "identifierUris[]", oneApp.IdentifierUris)
		if len(oneApp.IdentifierUris) == 0 {
			continue
		}

		identifierUrl := oneApp.IdentifierUris[0]
		log.Info("1 ApimProviderService.DiscoverApplications", "App.Name", oneApp.Name, "identifierUrl[0]", identifierUrl)

		apimServiceInfo, err := s.getApimServiceInfo(identifierUrl)
		if err != nil {
			log.Error("ApimProviderService.DiscoverApplications", "error calling getApimServiceInfo App.Name", oneApp.Name, "identifierUrl", identifierUrl, "err=", err)
			return nil, err
		}

		if apimServiceInfo.ResourceGroup == "" || apimServiceInfo.Name == "" {
			log.Info("3c ApimProviderService.DiscoverApplications ignoring app, no matching apim service found", "App.Name", oneApp.Name, "identifierUrl", identifierUrl)
			continue
		}

		log.Info("3d ApimProviderService.DiscoverApplications found apim service with matching identifierUrl", "App.Name", oneApp.Name, "identifierUrl", identifierUrl)
		apps = append(apps, orchestrator.ApplicationInfo{
			ObjectID:    oneApp.AppID,
			Name:        oneApp.Name,
			Description: oneApp.ID,
			Service:     identifierUrl,
		})
	}

	return apps, nil
}

func (s *ApimProviderService) GetPolicyInfo(appInfo orchestrator.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	serviceInfoAndRars, err := s.getResourceRolesForApi(appInfo)
	if err != nil {
		log.Error("ApimProviderService.GetPolicyInfo", "error calling getResourceRolesForApi App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return []hexapolicy.PolicyInfo{}, err
	}
	return providerscommon.BuildPolicies(serviceInfoAndRars.rarList), nil
}

func (s *ApimProviderService) SetPolicyInfo(appInfo orchestrator.ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (int, error) {
	serviceInfoAndRars, err := s.getResourceRolesForApi(appInfo)
	if err != nil {
		log.Error("ApimProviderService.SetPolicyInfo", "error calling getResourceRolesForApi App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return http.StatusBadGateway, err
	}

	rarUpdateList := providerscommon.CalcResourceActionRolesForUpdate(serviceInfoAndRars.rarList, policyInfos)
	serviceInfo := serviceInfoAndRars.serviceInfo
	for _, rar := range rarUpdateList {
		err := s.apimNamedValueSvc.UpdateResourceRole(serviceInfo, rar)
		if err != nil {
			return http.StatusBadGateway, err
		}
	}
	return http.StatusCreated, nil
}

type serviceAndRars struct {
	serviceInfo armmodel.ApimServiceInfo
	rarList     []providerscommon.ResourceActionRoles
}

func (s *ApimProviderService) getResourceRolesForApi(appInfo orchestrator.ApplicationInfo) (serviceAndRars, error) {
	serviceInfo, err := s.getApimServiceInfo(appInfo.Service)
	if err != nil {
		log.Error("ApimProviderService.SetPolicyInfo", "error calling getApimServiceInfo App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return serviceAndRars{}, err
	}

	log.Info("ApimProviderService.GetPolicyInfo", "apimServiceInfo", serviceInfo)

	rarList, err := s.apimNamedValueSvc.GetResourceRoles(serviceInfo)
	return serviceAndRars{
		serviceInfo: serviceInfo,
		rarList:     rarList,
	}, nil
}

func (s *ApimProviderService) getApimServiceInfo(identifierUrl string) (armmodel.ApimServiceInfo, error) {
	log.Info("ApimProviderService.getApimServiceInfo", "identifierUrl", identifierUrl)

	if identifierUrl == "" {
		errMsg := fmt.Sprintf("ApimProviderService.getApimServiceInfo identifierUrl is empty")
		log.Error(errMsg)
		return armmodel.ApimServiceInfo{}, errors.New(errMsg)
	}

	serviceInfo, err := s.armApimSvc.GetApimServiceInfo(identifierUrl)
	if err != nil {
		log.Error("ApimProviderService.getApimServiceInfo", "GetApimApiInfo err", err)
		return armmodel.ApimServiceInfo{}, err
	}

	return serviceInfo, nil
}

/*

func getApimProviderService(key []byte) (ApimProviderService, error) {
	var apimSvc ArmApimSvc
	var apimNamedValueSvc apimnv.ApimNamedValueSvc
	var azureClient azad.AzureClient

	if a.hasOverrides {
		apimSvc = a.armApimSvcOverride
		apimNamedValueSvc = a.apimNamedValueSvcOverride
		azureClient = a.azureClientOverride
	} else {
		factory, err := NewSvcFactory(key, nil)
		if err != nil {
			log.Error("ApimProvider.getApimService", "NewSvcFactory", "error=", err)
			return nil, err
		}

		apimSvc, _ = factory.NewApimSvc()
		apimNamedValueSvc = factory.NewApimNamedValueSvc()
		azureClient = azad.NewAzureClient(nil)
	}

	return NewApimProviderService(apimSvc, azureClient, apimNamedValueSvc), nil
}

*/
