package microsoftazure

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"net/http"
	"strings"
)

type AzureProvider struct {
	Http HTTPClient
}

func (g AzureProvider) Name() string {
	return "azure"
}

func (a AzureProvider) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo, err error) {
	key := info.Key
	if strings.EqualFold(info.Name, a.Name()) {
		if a.Http == nil {
			a.Http = &http.Client{} // todo - for testing, might be a better way?
		}
		azureClient := AzureClient{a.Http}
		found, _ := azureClient.GetWebApplications(key)
		apps = append(apps, found...)
	}
	return apps, err
}

func (a AzureProvider) GetPolicyInfo(integrationInfo provider.IntegrationInfo, applicationInfo provider.ApplicationInfo) ([]provider.PolicyInfo, error) {
	return []provider.PolicyInfo{}, nil
}

func (a AzureProvider) SetPolicyInfo(integrationInfo provider.IntegrationInfo, applicationInfo provider.ApplicationInfo, policyInfo provider.PolicyInfo) error {
	return nil
}
