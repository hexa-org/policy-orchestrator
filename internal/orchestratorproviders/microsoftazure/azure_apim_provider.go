package microsoftazure

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	log "golang.org/x/exp/slog"
	"net/http"
	"strings"
)

type AzureApimProvider struct {
	armApimSvcOverride  azapim.ArmApimSvc
	azureClientOverride azad.AzureClient
}

type AzureApimProviderOpt func(provider *AzureApimProvider)

func WithArmApimSvcOverride(armApimSvcOverride azapim.ArmApimSvc) func(provider *AzureApimProvider) {
	return func(provider *AzureApimProvider) {
		provider.armApimSvcOverride = armApimSvcOverride
	}
}

func WithAzureClientOverride(azureClientOverride azad.AzureClient) func(provider *AzureApimProvider) {
	return func(provider *AzureApimProvider) {
		provider.azureClientOverride = azureClientOverride
	}
}

func NewAzureApimProvider(opts ...AzureApimProviderOpt) *AzureApimProvider {
	provider := &AzureApimProvider{}
	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

func (a *AzureApimProvider) Name() string {
	return "azure"
}

func (a *AzureApimProvider) DiscoverApplications(integrationInfo orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	log.Info("ApimProvider.DiscoverApplications", "info.Name", integrationInfo.Name, "a.Name", a.Name())
	if !strings.EqualFold(integrationInfo.Name, a.Name()) {
		return []orchestrator.ApplicationInfo{}, err
	}

	service, err := a.getApimProviderService(integrationInfo.Key)
	if err != nil {
		log.Error("ApimProvider.GetPolicyInfo", "getApimProviderService err", err)
		return []orchestrator.ApplicationInfo{}, err
	}

	return service.DiscoverApplications(integrationInfo)

}

func (a *AzureApimProvider) GetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	service, err := a.getApimProviderService(integrationInfo.Key)
	if err != nil {
		log.Error("ApimProvider.GetPolicyInfo", "getApimProviderService err", err)
		return []policysupport.PolicyInfo{}, err
	}

	return service.GetPolicyInfo(applicationInfo)
}

func (a *AzureApimProvider) SetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	service, err := a.getApimProviderService(integrationInfo.Key)
	if err != nil {
		log.Error("ApimProvider.SetPolicyInfo", "getApimProviderService err", err)
		return http.StatusBadGateway, nil
	}

	return service.SetPolicyInfo(applicationInfo, policyInfos)
}

func (a *AzureApimProvider) getApimProviderService(key []byte) (*azapim.ApimProviderService, error) {
	armApimSvc, err := a.getApimSvc(key)
	if err != nil {
		return nil, err
	}
	return azapim.NewApimProviderService(armApimSvc, a.getAzureClient()), nil
}

func (a *AzureApimProvider) getApimSvc(key []byte) (azapim.ArmApimSvc, error) {
	if a.armApimSvcOverride != nil {
		return a.armApimSvcOverride, nil
	}

	factory, err := NewSvcFactory(key, nil)
	if err != nil {
		log.Error("ApimProvider.getApimService", "NewSvcFactory", "error=", err)
		return nil, err
	}

	apimSvc, err := factory.NewApimSvc()

	if err != nil {
		log.Error("ApimProvider.getApimService", "NewArmApimSvc", "err=", err)
		return nil, err
	}
	return apimSvc, nil
}

func (a *AzureApimProvider) getAzureClient() azad.AzureClient {
	if a.azureClientOverride != nil {
		return a.azureClientOverride
	}

	return azad.NewAzureClient(nil)
}
