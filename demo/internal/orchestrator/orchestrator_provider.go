package orchestrator

import (
	"fmt"
	"strings"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/sdk"
)

type providerBuilder struct {
	providerCache map[string]policyprovider.Provider
}

// var legacyProviders = map[string]Provider{
//	"google_cloud":      &googlecloud.GoogleProvider{},
//	"open_policy_agent": &openpolicyagent.OpaProvider{},
// }

func NewProviderBuilder() *providerBuilder {
	return &providerBuilder{providerCache: make(map[string]policyprovider.Provider)}
}

// legacyProviders["google_cloud"] = &googlecloud.GoogleProvider{}
// providers["azure"] = azarm.NewAzureApimProvider()
// providers["azure_apim"] = microsoftazure.NewAzureApimProvider()
// providers["amazon"] = &amazonwebservices.AmazonProvider{}
// providers["amazon"] = &awsapigw.AwsApiGatewayProvider{}
// providers["open_policy_agent"] = &openpolicyagent.OpaProvider{}

func MapSdkProviderName(legacyName string) string {
	switch strings.ToLower(legacyName) {
	case "azure", "azure_apim":
		return sdk.ProviderTypeAzure
	case "amazon":
		return sdk.ProviderTypeCognito
	case "google_cloud", "gcp":
		return sdk.ProviderTypeGoogleCloudIAP
	case "open_policy_agent":
		return sdk.ProviderTypeOpa
	case "noop":
		return "noop"
	}
	return legacyName
}

// AddProviders is primarily used in testing to allow a test provider to be directly injected rather than from the sdk integration providers
func (b *providerBuilder) AddProviders(cacheProviders map[string]policyprovider.Provider) {
	for k, v := range cacheProviders {
		b.providerCache[k] = v
	}
}

// GetAppsProvider returns a policyprovider.Provider that can be used to retrieve applications, as well as get and set policies
func (b *providerBuilder) GetAppsProvider(id string, providerType string, key []byte) (policyprovider.Provider, error) {

	provider, ok := b.providerCache[id]
	if ok {
		return provider, nil
	}

	info := policyprovider.IntegrationInfo{
		Name: MapSdkProviderName(providerType),
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
