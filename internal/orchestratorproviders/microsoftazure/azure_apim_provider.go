package microsoftazure

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	log "golang.org/x/exp/slog"
	"net/http"
	"strings"
)

type AzureApimProvider struct {
	//client              HTTPClient
	armApimSvcOverride  azureapim.ArmApimSvc
	azureClientOverride AzureClient
}

type AzureApimProviderOpt func(provider *AzureApimProvider)

func WithArmApimSvcOverride(armApimSvcOverride azureapim.ArmApimSvc) func(provider *AzureApimProvider) {
	return func(provider *AzureApimProvider) {
		provider.armApimSvcOverride = armApimSvcOverride
	}
}

func WithAzureClientOverride(azureClientOverride AzureClient) func(provider *AzureApimProvider) {
	return func(provider *AzureApimProvider) {
		provider.azureClientOverride = azureClientOverride
	}
}

//func WithAzureApimClient(clientOverride HTTPClient) func(provider *AzureApimProvider) {
//	return func(provider *AzureApimProvider) {
//		provider.client = clientOverride
//	}
//}

func NewAzureApimProvider(opts ...AzureApimProviderOpt) *AzureApimProvider {
	//provider := &AzureApimProvider{client: NewAzureClient(&http.Client{})}
	provider := &AzureApimProvider{}
	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

func (a *AzureApimProvider) Name() string {
	return "azure"
}

func (a *AzureApimProvider) DiscoverApplications(info orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	log.Info("ApimProvider.DiscoverApplications", "info.Name", info.Name, "a.Name", a.Name())
	if !strings.EqualFold(info.Name, a.Name()) {
		return apps, err
	}

	apimService, err := a.getApimSvc(info.Key)
	if err != nil {
		log.Error("ApimProvider.DiscoverApplications", "NewArmApimSvc err", err)
		return nil, err
	}

	aadClient := a.getAzureClient()
	azWebApps, err := aadClient.GetAzureApplications(info.Key)
	if err != nil {
		log.Error("ApimProvider.DiscoverApplications", "GetAzureApplications err", err)
		return nil, err
	}

	for _, oneApp := range azWebApps {
		log.Info("1 ApimProvider.DiscoverApplications", "App.Name", oneApp.Name, "identifierUris[]", oneApp.IdentifierUris)
		if len(oneApp.IdentifierUris) == 0 {
			continue
		}

		identifierUrl := oneApp.IdentifierUris[0]
		log.Info("1 ApimProvider.DiscoverApplications", "App.Name", oneApp.Name, "identifierUrl[0]", identifierUrl)

		apimServiceInfo, err := apimService.GetApimServiceInfo(identifierUrl)
		if err != nil {
			log.Error("ApimProvider.DiscoverApplications", "GetApimApiInfo err", err)
			return nil, err
		}

		if apimServiceInfo.ResourceGroup == "" || apimServiceInfo.Name == "" {
			log.Info("3c ApimProvider.DiscoverApplications ignoring app, no matching apim service with identifierUrl", "App.Name", oneApp.Name, "identifierUrl", identifierUrl)
			continue
		}

		log.Info("3d ApimProvider.DiscoverApplications found apim service with matching identifierUrl", "App.Name", oneApp.Name, "identifierUrl", identifierUrl)
		apps = append(apps, orchestrator.ApplicationInfo{
			ObjectID:    oneApp.AppID,
			Name:        oneApp.Name,
			Description: oneApp.ID,
			Service:     identifierUrl,
		})
	}

	return apps, err
}

func (a *AzureApimProvider) GetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	service, err := a.getApimProviderService(integrationInfo.Key)
	if err != nil {
		log.Error("ApimProvider.GetPolicyInfo", "getApimProviderService err", err)
		return []policysupport.PolicyInfo{}, err
	}

	return service.GetPolicyInfo(applicationInfo)
	/*identifierUrl := applicationInfo.Service
	log.Info("ApimProvider.GetPolicyInfo", "App.Name", applicationInfo.Name, "identifierUrl[0]", identifierUrl)

	if identifierUrl == "" {
		errMsg := fmt.Sprintf("ApimProvider.GetPolicyInfo identifierUrl not found. AppInfo.ID=%s AppInfo.Name=%s", applicationInfo.ObjectID, applicationInfo.Name)
		log.Error(errMsg)
		return []policysupport.PolicyInfo{}, errors.New(errMsg)
	}

	apimSvc, err := a.getApimSvc(integrationInfo.Key)
	if err != nil {
		log.Error("ApimProvider.GetPolicyInfo", "NewArmApimSvc err", err)
		return []policysupport.PolicyInfo{}, err
	}

	apimServiceInfo, err := apimSvc.GetApimServiceInfo(identifierUrl)
	if err != nil {
		log.Error("ApimProvider.GetPolicyInfo", "GetApimApiInfo err", err)
		return []policysupport.PolicyInfo{}, err
	}
	//var zeroApimServiceInfo armmodel.ApimServiceInfo
	//if apimServiceInfo == zeroApimServiceInfo {
	//	log.Info("3c ApimProvider", "GetApimApiInfo apimServiceInfo nil")
	//	return policies, nil
	//}

	log.Info("ApimProvider.GetPolicyInfo", "apimServiceInfo", apimServiceInfo)

	resourceActionRolesList, err := apimSvc.GetResourceRoles(apimServiceInfo)
	if err != nil {
		log.Error("ApimProvider.GetPolicyInfo", "GetResourceRoles err", err)
		return []policysupport.PolicyInfo{}, err
	}

	return buildPolicies(resourceActionRolesList), nil*/
}

func (a *AzureApimProvider) SetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	return http.StatusBadGateway, nil
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

func (a *AzureApimProvider) getApimProviderService(key []byte) (*azureapim.ApimProviderService, error) {
	armApimSvc, err := a.getApimSvc(key)
	if err != nil {
		return nil, err
	}
	return azureapim.NewApimProviderService(armApimSvc, a.getAzureClient()), nil
}
func (a *AzureApimProvider) getApimSvc(key []byte) (azureapim.ArmApimSvc, error) {
	if a.armApimSvcOverride != nil {
		return a.armApimSvcOverride, nil
	}

	factory, err := NewApimProviderSvcFactory(key, nil)
	if err != nil {
		log.Error("ApimProvider.getApimService", "NewApimProviderSvcFactory", "error=", err)
		return nil, err
	}

	apimService, err := factory.NewApimSvc()

	if err != nil {
		log.Error("ApimProvider.getApimService", "NewArmApimSvc", "err=", err)
		return nil, err
	}
	return apimService, nil
}

func (a *AzureApimProvider) getAzureClient() AzureClient {
	if a.azureClientOverride != nil {
		return a.azureClientOverride
	}

	return NewAzureClient(nil)
}
