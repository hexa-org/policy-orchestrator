package azuretestsupport

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"net/http"
	"net/url"
)

const LoginMicrosoftOnlineUrl = "https://login.microsoftonline.com"
const GraphApiBaseUrl = "https://graph.microsoft.com/v1.0"

type AzureHttpClient struct {
	AppId              string
	TenantId           string
	ServicePrincipalId string
	MockHttpClient     *testsupport.MockHTTPClient
}

func NewAzureHttpClient() *AzureHttpClient {
	return &AzureHttpClient{
		AppId:              AzureAppId,
		TenantId:           AzureTenantId,
		ServicePrincipalId: ServicePrincipalId,
		MockHttpClient:     testsupport.NewMockHTTPClient(),
	}
}

func (ac *AzureHttpClient) AzureClient() microsoftazure.AzureClient {
	client := microsoftazure.NewAzureClient(ac.MockHttpClient)
	return client
}
func (ac *AzureHttpClient) ErrorRequest(method string, url string, expStatus int, body []byte) {
	ac.MockHttpClient.AddRequest(method, url, expStatus, body)
}

func (ac *AzureHttpClient) TokenUrl() string {
	tokenUrl := fmt.Sprintf(`%s/%s/oauth2/v2.0/token`, LoginMicrosoftOnlineUrl, ac.TenantId)
	return tokenUrl
}
func (ac *AzureHttpClient) TokenRequest(token string) {
	respJson := fmt.Sprintf(`{"access_token":"%s"}`, token)
	ac.MockHttpClient.AddRequest(http.MethodPost, ac.TokenUrl(), http.StatusOK, []byte(respJson))
}
func (ac *AzureHttpClient) TokenCalled() bool {
	return ac.MockHttpClient.CalledWithStatus(http.MethodPost, ac.TokenUrl(), http.StatusOK)
}

func (ac *AzureHttpClient) GetWebApplicationsRequest(expResp string) {
	url := GraphApiBaseUrl + "/applications"
	resp := []byte(expResp)
	ac.MockHttpClient.AddRequest(http.MethodGet, url, http.StatusOK, resp)
}

func (ac *AzureHttpClient) GetServicePrincipalsUrl() string {
	url := fmt.Sprintf(`%s/servicePrincipals?$search="appId:%s"`, GraphApiBaseUrl, ac.AppId)
	return url
}

func (ac *AzureHttpClient) GetServicePrincipalsRequest() {
	resp := []byte(ServicePrincipalsRespJson)
	ac.MockHttpClient.AddRequest(http.MethodGet, ac.GetServicePrincipalsUrl(), http.StatusOK, resp)
}

func (ac *AzureHttpClient) GetUserInfoFromPrincipalIdUrl(principalId string) string {
	url := fmt.Sprintf(`%s/users/%s`, GraphApiBaseUrl, principalId)
	return url
}

func (ac *AzureHttpClient) GetUserInfoFromPrincipalIdRequest(principalId string) {
	azUser := microsoftazure.AzureUser{
		PrincipalId: principalId,
		Email:       policytestsupport.MakeEmail(principalId),
	}
	resp, _ := json.Marshal(azUser)
	ac.MockHttpClient.AddRequest(http.MethodGet, ac.GetUserInfoFromPrincipalIdUrl(principalId), http.StatusOK, resp)
}

func (ac *AzureHttpClient) GetPrincipalIdFromEmailUrl(principalId string) string {
	email := policytestsupport.MakeEmail(principalId)
	url := fmt.Sprintf("%s/users?$select=id,mail&$filter=mail%%20eq%%20%%27%s%%27", GraphApiBaseUrl, url.QueryEscape(email))
	return url
}

func (ac *AzureHttpClient) GetPrincipalIdFromEmailRequest(principalId string) {
	azUser := microsoftazure.AzureUsers{
		List: []microsoftazure.AzureUser{{
			PrincipalId: principalId,
			Email:       policytestsupport.MakeEmail(principalId),
		}},
	}

	resp, _ := json.Marshal(azUser)
	ac.MockHttpClient.AddRequest(http.MethodGet, ac.GetPrincipalIdFromEmailUrl(principalId), http.StatusOK, resp)
}

func (ac *AzureHttpClient) AppRoleAssignmentsUrl() string {
	appRoleAssignmentsUrl := fmt.Sprintf(`%s/servicePrincipals/%s/appRoleAssignedTo`, GraphApiBaseUrl, ac.ServicePrincipalId)
	return appRoleAssignmentsUrl
}

func (ac *AzureHttpClient) GetAppRoleAssignmentsRequest(appRoleAssignments []microsoftazure.AzureAppRoleAssignment) {
	assignmentList := microsoftazure.AzureAppRoleAssignments{
		List: appRoleAssignments,
	}

	resp, _ := json.Marshal(assignmentList)
	ac.MockHttpClient.AddRequest(http.MethodGet, ac.AppRoleAssignmentsUrl(), http.StatusOK, resp)
}

func (ac *AzureHttpClient) GetAppRoleAssignmentsCalled() bool {
	return ac.MockHttpClient.CalledWithStatus(http.MethodGet, ac.AppRoleAssignmentsUrl(), http.StatusOK)
}

func (ac *AzureHttpClient) PostAppRoleAssignmentsRequest() {
	ac.MockHttpClient.AddRequest(http.MethodPost, ac.AppRoleAssignmentsUrl(), http.StatusCreated, nil)
}

func (ac *AzureHttpClient) PostAppRoleAssignmentsCalled() bool {
	return ac.MockHttpClient.CalledWithStatus(http.MethodPost, ac.AppRoleAssignmentsUrl(), http.StatusCreated)
}

func (ac *AzureHttpClient) DeleteAppRoleAssignmentsRequest(toDelete []microsoftazure.AzureAppRoleAssignment) {
	for _, ara := range toDelete {
		deleteUrl := ac.AppRoleAssignmentsUrl() + "/" + ara.ID
		ac.MockHttpClient.AddRequest(http.MethodDelete, deleteUrl, http.StatusNoContent, nil)
	}

}

func (ac *AzureHttpClient) DeleteAppRoleAssignmentsCalled(deleted microsoftazure.AzureAppRoleAssignment) bool {
	deleteUrl := ac.AppRoleAssignmentsUrl() + "/" + deleted.ID
	return ac.MockHttpClient.CalledWithStatus(http.MethodDelete, deleteUrl, http.StatusNoContent)
}
