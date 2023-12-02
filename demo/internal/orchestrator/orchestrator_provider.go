package orchestrator

import (
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2/apps"
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2/policy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policyprovider"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp"
	log "golang.org/x/exp/slog"
	"net/http"
)

type OrchestrationProvider struct {
	idp         string
	policyStore string
	service     policyprovider.ProviderService
}

func NewOrchestrationProvider[R rar.ResourceActionRolesMapper](aIdp apps.Idp, policyStoreOpt policy.PolicyStore[R]) (*OrchestrationProvider, error) {
	// idpCredentials []byte, policyStoreCredentials []byte
	//tableInfo, err := dynamodbpolicystore.NewSimpleTableInfo(awsPolicyStoreTableName, resourcePolicyItem{})
	//policyStoreSvc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, policyStoreOpt.key)
	policyStoreSvc, err := policyStoreOpt.Provider()
	if err != nil {
		log.Error("NewOrchestrationProvider",
			"msg", "failed to create dynamodbpolicystore.PolicyStoreSvc",
			"error", err)
		return nil, err
	}

	//appInfoSvc, err := cognitoidp.NewAppInfoSvc(idpOpt.key)
	appInfoSvc, err := aIdp.Provider()
	if err != nil {
		log.Error("NewAwsApiGatewayProviderV2",
			"msg", "failed to create cognitoidp.AppInfoSvc",
			"error", err)
		return nil, err
	}

	service := policyprovider.NewProviderService[R](appInfoSvc, policyStoreSvc)
	provider := &OrchestrationProvider{
		service: service,
	}
	return provider, nil
}

func (a *OrchestrationProvider) Name() string {
	return "amazon"
}

func (a *OrchestrationProvider) DiscoverApplications(_ IntegrationInfo) ([]ApplicationInfo, error) {
	discoveredApps, err := a.service.DiscoverApplications()
	if err != nil {
		return nil, err
	}

	retApps := make([]ApplicationInfo, 0)
	for _, oneApp := range discoveredApps {
		retApps = append(retApps, toApplicationInfo(oneApp))
	}

	return retApps, nil

}

func (a *OrchestrationProvider) GetPolicyInfo(_ IntegrationInfo, applicationInfo ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	idpAppInfo := toIdpAppInfo(applicationInfo)
	return a.service.GetPolicyInfo(idpAppInfo)
}

func (a *OrchestrationProvider) SetPolicyInfo(_ IntegrationInfo, applicationInfo ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (status int, foundErr error) {
	log.Info("SetPolicyInfo", "msg", "BEGIN",
		"applicationInfo.ObjectID", applicationInfo.ObjectID,
		"Name", applicationInfo.Name,
		"Description", applicationInfo.Description,
		"Service", applicationInfo.Service)

	idpAppInfo := toIdpAppInfo(applicationInfo)
	err := a.service.SetPolicyInfo(idpAppInfo, policyInfos)
	log.Info("SetPolicyInfo", "msg", "Finished calling service.SetPolicyInfo")

	if err != nil {
		log.Error("SetPolicyInfo", "msg", "error calling service.SetPolicyInfo", "error", err)
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func toApplicationInfo(anApp idp.AppInfo) ApplicationInfo {
	rsApp := (anApp).(cognitoidp.ResourceServerAppInfo)
	return ApplicationInfo{
		ObjectID:    rsApp.Id(),
		Name:        rsApp.Name(),
		Description: rsApp.DisplayName(),
		Service:     rsApp.Identifier(),
	}
}

func toIdpAppInfo(applicationInfo ApplicationInfo) idp.AppInfo {
	return cognitoidp.NewResourceServerAppInfo(applicationInfo.ObjectID, applicationInfo.Name, applicationInfo.Description, applicationInfo.Service)
}
