package orchestrator

import (
	"fmt"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/migrationSupport"
)

type ProviderBuilder struct {
	providerCache map[string]policyprovider.Provider
}

// var legacyProviders = map[string]Provider{
//	"google_cloud":      &googlecloud.GoogleProvider{},
//	"open_policy_agent": &openpolicyagent.OpaProvider{},
// }

func NewProviderBuilder() *ProviderBuilder {
	return &ProviderBuilder{providerCache: make(map[string]policyprovider.Provider)}
}

// legacyProviders["google_cloud"] = &googlecloud.GoogleProvider{}
// providers["azure"] = azarm.NewAzureApimProvider()
// providers["azure_apim"] = microsoftazure.NewAzureApimProvider()
// providers["amazon"] = &amazonwebservices.AmazonProvider{}
// providers["amazon"] = &awsapigw.AwsApiGatewayProvider{}
// providers["open_policy_agent"] = &openpolicyagent.OpaProvider{}

// AddProviders is primarily used in testing to allow a test provider to be directly injected rather than from the sdk integration providers
func (b *ProviderBuilder) AddProviders(cacheProviders map[string]policyprovider.Provider) {
	for k, v := range cacheProviders {
		b.providerCache[k] = v
	}
}

// GetAppsProvider returns a policyprovider.Provider that can be used to retrieve applications, as well as get and set policies
func (b *ProviderBuilder) GetAppsProvider(id string, providerType string, key []byte) (policyprovider.Provider, error) {
	if b.providerCache == nil {
		fmt.Print("... provider cache is nil! ")
		return nil, nil
	}
	provider, ok := b.providerCache[id]
	if ok {
		return provider, nil
	}

	info := policyprovider.IntegrationInfo{
		Name: migrationSupport.MapSdkProviderName(providerType),
		Key:  key,
	}

	var err error
	integration, err := sdk.OpenIntegration(sdk.WithIntegrationInfo(info))
	if err != nil {
		return nil, fmt.Errorf("failed to GetOrchestrationProvider. no such provider found %s", providerType)
	}

	b.providerCache[id] = integration.GetProvider()

	return integration.GetProvider(), nil

}
