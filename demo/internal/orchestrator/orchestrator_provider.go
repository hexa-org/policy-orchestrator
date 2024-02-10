package orchestrator

import (
	"fmt"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	serviceprovider "github.com/hexa-org/policy-orchestrator/sdk/core/policyprovider"

	"net/http"

	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp"
	logger "golang.org/x/exp/slog"
)

type providerBuilder struct {
	legacyProviders map[string]policyprovider.Provider
}

// var legacyProviders = map[string]Provider{
//	"google_cloud":      &googlecloud.GoogleProvider{},
//	"open_policy_agent": &openpolicyagent.OpaProvider{},
// }

func NewProviderBuilder(legacyProviders map[string]policyprovider.Provider) *providerBuilder {
	return &providerBuilder{legacyProviders: legacyProviders}
}

// legacyProviders["google_cloud"] = &googlecloud.GoogleProvider{}
// providers["azure"] = azarm.NewAzureApimProvider()
// providers["azure_apim"] = microsoftazure.NewAzureApimProvider()
// providers["amazon"] = &amazonwebservices.AmazonProvider{}
// providers["amazon"] = &awsapigw.AwsApiGatewayProvider{}
// providers["open_policy_agent"] = &openpolicyagent.OpaProvider{}

var idpMap = map[string]func(key []byte) providersV2.Idp{
	"amazon": func(key []byte) providersV2.Idp {
		return providersV2.NewCognitoIdp(key)
	},
	"azure": func(key []byte) providersV2.Idp {
		return providersV2.NewApimAppProvider(key)
	},
}

func (b *providerBuilder) GetAppsProvider(provider string, key []byte) (policyprovider.Provider, error) {
	legacyProvider, found := b.legacyProviders[provider]
	if found {
		return legacyProvider, nil
	}

	appFunc, found := idpMap[provider]
	if !found {
		return nil, fmt.Errorf("failed to GetOrchestrationProvider. no such provider found %s", provider)
	}
	p, err := NewOrchestrationProvider(provider, appFunc(key), providersV2.NewEmptyPolicyStore())
	return p, err
}

/*
func (b *providerBuilder) GetOrchestrationProvider(provider string, appsKey []byte, policyStoreOpt policy.PolicyStore[any]) (Provider, error) {
	legacyProvider, found := b.legacyProviders[provider]
	if found {
		return legacyProvider, nil
	}

	appFunc, found := idpMap[provider]
	if !found {
		return nil, fmt.Errorf("failed to GetOrchestrationProvider. no such provider found %s", provider)
	}
	p, err := NewOrchestrationProvider(appFunc(appsKey), policyStoreOpt)
	return p, err
}
*/

type OrchestrationProvider struct {
	name        string
	idp         string
	policyStore string
	service     serviceprovider.ProviderService
}

func NewOrchestrationProvider[R any](name string, aIdp providersV2.Idp, policyStoreOpt providersV2.PolicyStore[R]) (*OrchestrationProvider, error) {
	// idpCredentials []byte, policyStoreCredentials []byte
	// tableInfo, err := dynamodbpolicystore.NewSimpleTableInfo(awsPolicyStoreTableName, resourcePolicyItem{})
	// policyStoreSvc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, policyStoreOpt.key)
	policyStoreSvc, err := policyStoreOpt.Provider()
	if err != nil {
		logger.Error("NewOrchestrationProvider",
			"msg", "failed to create dynamodbpolicystore.PolicyStoreSvc",
			"error", err)
		return nil, err
	}

	// appInfoSvc, err := cognitoidp.NewAppInfoSvc(idpOpt.key)
	appInfoSvc, err := aIdp.Provider()
	if err != nil {
		logger.Error("NewAwsApiGatewayProviderV2",
			"msg", "failed to create cognitoidp.AppInfoSvc",
			"error", err)
		return nil, err
	}

	service := serviceprovider.NewProviderService[R](appInfoSvc, policyStoreSvc)
	provider := &OrchestrationProvider{
		name:    name,
		service: service,
	}
	return provider, nil
}

func (a *OrchestrationProvider) Name() string {
	return a.name
}

func (a *OrchestrationProvider) DiscoverApplications(_ policyprovider.IntegrationInfo) ([]policyprovider.ApplicationInfo, error) {
	discoveredApps, err := a.service.DiscoverApplications()
	if err != nil {
		return nil, err
	}

	retApps := make([]policyprovider.ApplicationInfo, 0)
	for _, oneApp := range discoveredApps {
		logger.Debug("DiscoverApplications", "id", oneApp.Id(), "Name", oneApp.Name(), "Display", oneApp.DisplayName(), "Type", oneApp.Type())
		retApps = append(retApps, toApplicationInfo(oneApp))
	}

	return retApps, nil

}

func (a *OrchestrationProvider) GetPolicyInfo(_ policyprovider.IntegrationInfo, applicationInfo policyprovider.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	idpAppInfo := toIdpAppInfo(applicationInfo)
	return a.service.GetPolicyInfo(idpAppInfo)
}

func (a *OrchestrationProvider) SetPolicyInfo(_ policyprovider.IntegrationInfo, applicationInfo policyprovider.ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (status int, foundErr error) {
	logger.Info("SetPolicyInfo", "msg", "BEGIN",
		"applicationInfo.ObjectID", applicationInfo.ObjectID,
		"Name", applicationInfo.Name,
		"Description", applicationInfo.Description,
		"Service", applicationInfo.Service)

	idpAppInfo := toIdpAppInfo(applicationInfo)
	err := a.service.SetPolicyInfo(idpAppInfo, policyInfos)
	logger.Info("SetPolicyInfo", "msg", "Finished calling service.SetPolicyInfo")

	if err != nil {
		logger.Error("SetPolicyInfo", "msg", "error calling service.SetPolicyInfo", "error", err)
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

// toApplicationInfo - convert sdk ResourceServerAppInfo to
// demo apps ApplicationInfo
func toApplicationInfo(anApp idp.AppInfo) policyprovider.ApplicationInfo {
	/*rsApp := (anApp).(cognitoidp.ResourceServerAppInfo)
	  return ApplicationInfo{
	  	ObjectID:    rsApp.Id(),
	  	Name:        rsApp.Name(),
	  	Description: rsApp.DisplayName(),
	  	Service:     rsApp.Identifier(),
	  }*/
	return policyprovider.ApplicationInfo{
		ObjectID:    anApp.Id(),
		Name:        anApp.Name(),
		Description: anApp.DisplayName(),
		Service:     anApp.Type(),
	}
}

// toIdpAppInfo - convert demo apps ApplicationInfo to sdk ResourceServerAppInfo
func toIdpAppInfo(applicationInfo policyprovider.ApplicationInfo) idp.AppInfo {
	return cognitoidp.NewResourceServerAppInfo(applicationInfo.ObjectID, applicationInfo.Name, applicationInfo.Description, applicationInfo.Service)
}
