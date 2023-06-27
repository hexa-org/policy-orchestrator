package azapim

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	log "golang.org/x/exp/slog"
	"net/http"
)

type ApimProviderService struct {
	armApimSvc  ArmApimSvc
	azureClient azad.AzureClient
}

func NewApimProviderService(armApimSvc ArmApimSvc, azureClient azad.AzureClient) *ApimProviderService {

	return &ApimProviderService{armApimSvc: armApimSvc, azureClient: azureClient}
}

func (aps *ApimProviderService) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	azWebApps, err := aps.azureClient.GetAzureApplications(info.Key)
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

		apimServiceInfo, err := aps.getApimServiceInfo(identifierUrl)
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

func (aps *ApimProviderService) GetPolicyInfo(appInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	serviceInfoAndRars, err := aps.getResourceRolesForApi(appInfo)
	if err != nil {
		log.Error("ApimProviderService.GetPolicyInfo", "error calling getResourceRolesForApi App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return []policysupport.PolicyInfo{}, err
	}
	return providerscommon.BuildPolicies(serviceInfoAndRars.rarList), nil
}

func (aps *ApimProviderService) SetPolicyInfo(appInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	serviceInfoAndRars, err := aps.getResourceRolesForApi(appInfo)
	if err != nil {
		log.Error("ApimProviderService.SetPolicyInfo", "error calling getResourceRolesForApi App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return http.StatusBadGateway, err
	}

	rarUpdateList := providerscommon.CalcResourceActionRolesForUpdate(serviceInfoAndRars.rarList, policyInfos)
	serviceInfo := serviceInfoAndRars.serviceInfo
	for _, rar := range rarUpdateList {
		err := aps.armApimSvc.UpdateResourceRole(serviceInfo, rar)
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

func (aps *ApimProviderService) getResourceRolesForApi(appInfo orchestrator.ApplicationInfo) (serviceAndRars, error) {
	serviceInfo, err := aps.getApimServiceInfo(appInfo.Service)
	if err != nil {
		log.Error("ApimProviderService.SetPolicyInfo", "error calling getApimServiceInfo App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return serviceAndRars{}, err
	}

	log.Info("ApimProviderService.GetPolicyInfo", "apimServiceInfo", serviceInfo)

	rarList, err := aps.armApimSvc.GetResourceRoles(serviceInfo)
	return serviceAndRars{
		serviceInfo: serviceInfo,
		rarList:     rarList,
	}, nil
}

func (aps *ApimProviderService) getApimServiceInfo(identifierUrl string) (armmodel.ApimServiceInfo, error) {
	log.Info("ApimProviderService.getApimServiceInfo", "identifierUrl", identifierUrl)

	if identifierUrl == "" {
		errMsg := fmt.Sprintf("ApimProviderService.getApimServiceInfo identifierUrl is empty")
		log.Error(errMsg)
		return armmodel.ApimServiceInfo{}, errors.New(errMsg)
	}

	serviceInfo, err := aps.armApimSvc.GetApimServiceInfo(identifierUrl)
	if err != nil {
		log.Error("ApimProviderService.GetPolicyInfo", "GetApimApiInfo err", err)
		return armmodel.ApimServiceInfo{}, err
	}

	return serviceInfo, nil
}
