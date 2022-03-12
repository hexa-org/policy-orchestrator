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

func (a *AzureProvider) Name() string {
	return "azure"
}

func (a *AzureProvider) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo, err error) {
	key := info.Key
	if strings.EqualFold(info.Name, a.Name()) {
		a.ensureClientIsAvailable()
		azureClient := AzureClient{a.Http}
		found, _ := azureClient.GetWebApplications(key)
		apps = append(apps, found...)
	}
	return apps, err
}

func (a *AzureProvider) GetPolicyInfo(integrationInfo provider.IntegrationInfo, applicationInfo provider.ApplicationInfo) ([]provider.PolicyInfo, error) {
	key := integrationInfo.Key
	var policies []provider.PolicyInfo
	a.ensureClientIsAvailable()
	azureClient := AzureClient{a.Http}
	principal, _ := azureClient.GetServicePrincipals(key, applicationInfo.Description) // todo - description is named poorly
	assignments, _ := azureClient.GetAppRoleAssignedTo(key, principal.List[0].ID)

	var appRoleId string
	var principalIdsAndDisplayNames []string
	var resourceIdAndDisplayName string

	for _, assignment := range assignments.List {
		appRoleId = assignment.AppRoleId
		resourceIdAndDisplayName = fmt.Sprintf("%s:%s", assignment.ResourceId, assignment.ResourceDisplayName)
		principalIdsAndDisplayNames = append(principalIdsAndDisplayNames, fmt.Sprintf("%s:%s", assignment.PrincipalId, assignment.PrincipalDisplayName))
	}

	policies = append(policies, provider.PolicyInfo{
		Version: "0.2",
		Action:  appRoleId,
		Subject: provider.SubjectInfo{AuthenticatedUsers: principalIdsAndDisplayNames},
		Object:  provider.ObjectInfo{Resources: []string{resourceIdAndDisplayName}},
	})

	return policies, nil
}

func (a *AzureProvider) SetPolicyInfo(integrationInfo provider.IntegrationInfo, applicationInfo provider.ApplicationInfo, policyInfos []provider.PolicyInfo) error {
	key := integrationInfo.Key
	a.ensureClientIsAvailable()
	azureClient := AzureClient{a.Http}
	principal, _ := azureClient.GetServicePrincipals(key, applicationInfo.Description) // todo - description is named poorly
	for _, policyInfo := range policyInfos {
		var assignments []AzureAppRoleAssignment
		resources := policyInfo.Object.Resources[0]
		for _, user := range policyInfo.Subject.AuthenticatedUsers {
			assignments = append(assignments, AzureAppRoleAssignment{
				AppRoleId:   policyInfo.Action,
				PrincipalId: strings.Split(user, ":")[0],
				ResourceId:  strings.Split(resources, ":")[0],
			})
		}
		err := azureClient.SetAppRoleAssignedTo(key, principal.List[0].ID, assignments)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AzureProvider) ensureClientIsAvailable() {
	if a.Http == nil {
		a.Http = &http.Client{} // todo - for testing, might be a better way?
	}
}
