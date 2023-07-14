package azarm

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	log "golang.org/x/exp/slog"
	"net/http"
	"strings"
)

type AzureApimProvider struct {
	armApimSvcOverride        azapim.ArmApimSvc
	azureClientOverride       azad.AzureClient
	apimNamedValueSvcOverride apimnv.ApimNamedValueSvc
	hasOverrides              bool
}

type AzureApimProviderOpt func(provider *AzureApimProvider)

func WithArmApimSvcOverride(armApimSvcOverride azapim.ArmApimSvc) AzureApimProviderOpt {
	return func(provider *AzureApimProvider) {
		provider.armApimSvcOverride = armApimSvcOverride
		provider.hasOverrides = true
	}
}

func WithApimNamedValueSvcOverride(apimNamedValueSvcOverride apimnv.ApimNamedValueSvc) AzureApimProviderOpt {
	return func(provider *AzureApimProvider) {
		provider.apimNamedValueSvcOverride = apimNamedValueSvcOverride
		provider.hasOverrides = true
	}
}

func WithAzureClientOverride(azureClientOverride azad.AzureClient) AzureApimProviderOpt {
	return func(provider *AzureApimProvider) {
		provider.azureClientOverride = azureClientOverride
		provider.hasOverrides = true
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

func (a *AzureApimProvider) getApimProviderService(key []byte) (*ApimProviderService, error) {
	var armApimSvc azapim.ArmApimSvc
	var apimNamedValueSvc apimnv.ApimNamedValueSvc
	var azureClient azad.AzureClient

	if a.hasOverrides {
		armApimSvc = a.armApimSvcOverride
		apimNamedValueSvc = a.apimNamedValueSvcOverride
		azureClient = a.azureClientOverride
	} else {
		factory, err := azapim.NewSvcFactory(key, nil)
		if err != nil {
			log.Error("ApimProvider.getApimService", "NewSvcFactory", "error=", err)
			return nil, err
		}

		armApimSvc, _ = factory.NewApimSvc()
		apimNamedValueSvc = factory.NewApimNamedValueSvc()
		azureClient = azad.NewAzureClient(nil)
	}

	return NewApimProviderService(armApimSvc, azureClient, apimNamedValueSvc), nil
}
