package microsoftazure

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"net/http"
	"strings"
)

type AzureProvider struct {
	HttpClientOverride HTTPClient
}

func (a *AzureProvider) Name() string {
	return "azure"
}

func (a *AzureProvider) DiscoverApplications(info orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	if !strings.EqualFold(info.Name, a.Name()) {
		return apps, err
	}

	key := info.Key
	client := a.getHttpClient()
	azureClient := AzureClient{client}
	found, _ := azureClient.GetWebApplications(key)
	apps = append(apps, found...)
	return apps, err
}

func (a *AzureProvider) GetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	key := integrationInfo.Key
	var policies []policysupport.PolicyInfo
	client := a.getHttpClient()
	azureClient := AzureClient{client}
	principal, _ := azureClient.GetServicePrincipals(key, applicationInfo.Description) // todo - description is named poorly
	assignments, _ := azureClient.GetAppRoleAssignedTo(key, principal.List[0].ID)

	appRoleId := fmt.Sprintf("azure:%s", assignments.List[0].AppRoleId)

	users := workflowsupport.ProcessAsync[string, AzureAppRoleAssignment](assignments.List, func(assignment AzureAppRoleAssignment) (string, error) {
		user, _ := azureClient.GetUserInfoFromPrincipalId(key, assignment.PrincipalId)

		if user.Email == "" {
			return "", errors.New("no email")
		}

		return fmt.Sprintf("user:%s", user.Email), nil
	})

	policies = append(policies, policysupport.PolicyInfo{
		Meta:    policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{{appRoleId}},
		Subject: policysupport.SubjectInfo{Members: users},
		Object: policysupport.ObjectInfo{
			ResourceID: applicationInfo.ObjectID,
		},
	})

	return policies, nil
}

func (a *AzureProvider) SetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	validate := validator.New() // todo - move this up?
	errApp := validate.Struct(applicationInfo)
	if errApp != nil {
		return http.StatusInternalServerError, errApp
	}
	errPolicies := validate.Var(policyInfos, "omitempty,dive")
	if errPolicies != nil {
		return http.StatusInternalServerError, errPolicies
	}

	key := integrationInfo.Key
	client := a.getHttpClient()
	azureClient := AzureClient{client}
	principal, _ := azureClient.GetServicePrincipals(key, applicationInfo.Description) // todo - description is named poorly
	for _, policyInfo := range policyInfos {
		var assignments []AzureAppRoleAssignment
		for _, user := range policyInfo.Subject.Members {
			principalId, _ := azureClient.GetPrincipalIdFromEmail(key, strings.Split(user, ":")[1])
			if principalId == "" {
				continue
			}
			assignments = append(assignments, AzureAppRoleAssignment{
				AppRoleId:   strings.TrimPrefix(policyInfo.Actions[0].ActionUri, "azure:"),
				PrincipalId: principalId,
				ResourceId:  strings.Split(policyInfo.Object.ResourceID, ":")[0],
			})
		}
		err := azureClient.SetAppRoleAssignedTo(key, principal.List[0].ID, assignments)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusCreated, nil
}

func (a *AzureProvider) getHttpClient() HTTPClient {
	if a.HttpClientOverride != nil {
		return a.HttpClientOverride
	}
	return &http.Client{}
}
