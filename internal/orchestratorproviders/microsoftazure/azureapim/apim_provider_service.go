package azureapim

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	log "golang.org/x/exp/slog"
)

type ApimProviderService struct {
	armApimSvc  ArmApimSvc
	azureClient microsoftazure.AzureClient
}

func NewApimProviderService(armApimSvc ArmApimSvc, azureClient microsoftazure.AzureClient) *ApimProviderService {
	return &ApimProviderService{armApimSvc: armApimSvc, azureClient: azureClient}
}

func (aps *ApimProviderService) GetPolicyInfo(appInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	serviceInfo, err := aps.getApimServiceInfo(appInfo)
	if err != nil {
		return []policysupport.PolicyInfo{}, err
	}

	log.Info("ApimProviderService.GetPolicyInfo", "apimServiceInfo", serviceInfo)

	resourceActionRolesList, err := aps.armApimSvc.GetResourceRoles(serviceInfo)
	if err != nil {
		log.Error("ApimProviderService.GetPolicyInfo", "GetResourceRoles err", err)
		return []policysupport.PolicyInfo{}, err
	}

	return buildPolicies(resourceActionRolesList), nil
}

func (aps *ApimProviderService) getApimServiceInfo(appInfo orchestrator.ApplicationInfo) (armmodel.ApimServiceInfo, error) {
	log.Info("ApimProviderService.getApimServiceInfo", "App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service)

	if appInfo.Service == "" {
		errMsg := fmt.Sprintf("ApimProviderService.GetPolicyInfo identifierUrl not found. AppInfo.ID=%s AppInfo.Name=%s", appInfo.ObjectID, appInfo.Name)
		log.Error(errMsg)
		return armmodel.ApimServiceInfo{}, errors.New(errMsg)
	}

	identifierUrl := appInfo.Service

	//apimSvc, err := a.getApimSvc(integrationInfo.Key)
	//if err != nil {
	//	log.Error("ApimProvider.GetPolicyInfo", "NewArmApimSvc err", err)
	//	return []policysupport.PolicyInfo{}, err
	//}

	serviceInfo, err := aps.armApimSvc.GetApimServiceInfo(identifierUrl)
	if err != nil {
		log.Error("ApimProviderService.GetPolicyInfo", "GetApimApiInfo err", err)
		return armmodel.ApimServiceInfo{}, err
	}

	return serviceInfo, nil
}

func buildPolicies(resourceActionRolesList []armmodel.ResourceActionRoles) []policysupport.PolicyInfo {
	policies := make([]policysupport.PolicyInfo, 0)
	for _, one := range resourceActionRolesList {
		policies = append(policies, policysupport.PolicyInfo{
			Meta:    policysupport.MetaInfo{Version: "0.5"},
			Actions: []policysupport.ActionInfo{{one.Action}},
			Subject: policysupport.SubjectInfo{Members: one.Roles},
			Object:  policysupport.ObjectInfo{ResourceID: one.Resource},
		})
	}
	return policies
}
