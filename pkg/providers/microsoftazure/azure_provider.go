package microsoftazure

import (
	"fmt"
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
	key := integrationInfo.Key
	var policies []provider.PolicyInfo
	if a.Http == nil {
		a.Http = &http.Client{}
	}
	azureClient := AzureClient{a.Http}
	principal, _ := azureClient.GetServicePrincipals(key, applicationInfo.Description)
	assignments, _ := azureClient.GetAppRoleAssignedTo(key, principal.List[0].ID)
	for _, a := range assignments.List {
		policies = append(policies, provider.PolicyInfo{
			Version: "0.2",
			Action:  a.AppRoleId,
			Subject: provider.SubjectInfo{AuthenticatedUsers: []string{fmt.Sprintf("%s:%s", a.PrincipalId, a.PrincipalDisplayName)}},
			Object:  provider.ObjectInfo{Resources: []string{fmt.Sprintf("%s:%s", a.ResourceId, a.ResourceDisplayName)}},
		})
	}
	return policies, nil
}

func (a AzureProvider) SetPolicyInfo(integrationInfo provider.IntegrationInfo, applicationInfo provider.ApplicationInfo, policyInfo provider.PolicyInfo) error {
	return nil
}
