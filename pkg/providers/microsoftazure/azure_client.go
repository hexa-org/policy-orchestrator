package microsoftazure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"io"
	"log"
	"net/http"
	"strings"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type AzureClient struct {
	HttpClient HTTPClient
}

type AzureKey struct {
	AppId        string `json:"appId"`
	Secret       string `json:"secret"`
	Tenant       string `json:"tenant"`
	Subscription string `json:"subscription"`
}

type AzureAccessToken struct {
	Token string `json:"access_token"`
}

type azureWebApps struct {
	List []azureWebApp `json:"value"`
}

type azureWebApp struct {
	ID    string       `json:"id"`
	AppID string       `json:"appId"`
	Name  string       `json:"displayName"`
	Web   azureWebInfo `json:"web"`
}

type azureWebInfo struct {
	HomePageUrl string `json:"homePageUrl"`
}
type AzureServicePrincipals struct {
	List []azureServicePrincipal `json:"value"`
}

type azureServicePrincipal struct {
	ID string `json:"id"`
}

type AzureAppRoleAssignments struct {
	List []azureAppRoleAssignment `json:"value"`
}

type azureAppRoleAssignment struct {
	ID                   string `json:"id"`
	AppRoleId            string `json:"appRoleId"`
	PrincipalDisplayName string `json:"principalDisplayName"`
	PrincipalId          string `json:"principalId"`
	PrincipalType        string `json:"principalType"`
	ResourceDisplayName  string `json:"resourceDisplayName"`
	ResourceId           string `json:"resourceId"`
}

func (c *AzureClient) GetWebApplications(key []byte) ([]provider.ApplicationInfo, error) {

	request, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/applications", nil)
	get, err := c.azureRequest(key, request)
	if err != nil {
		log.Println("Unable to get azure web applications.")
		return []provider.ApplicationInfo{}, err
	}

	var webapps azureWebApps
	if err = json.NewDecoder(get.Body).Decode(&webapps); err != nil {
		log.Println("Unable to decode azure web app response.")
		return []provider.ApplicationInfo{}, err
	}

	var apps []provider.ApplicationInfo
	for _, app := range webapps.List {
		log.Printf("Found azure app service web app %s.\n", app.Name)
		apps = append(apps, provider.ApplicationInfo{ObjectID: app.ID, Name: app.Name, Description: app.AppID})
	}
	return apps, err
}

func (c *AzureClient) GetPolicy(key []byte) ([]provider.PolicyInfo, error) {

	return nil, nil
}

func (c *AzureClient) SetPolicy(key []byte, policy provider.PolicyInfo) error {
	return nil
}

func (c *AzureClient) GetServicePrincipals(key []byte, appId string) (AzureServicePrincipals, error) {

	filter := fmt.Sprintf("$search=\"appId:%s\"", appId)
	urlWithFilter := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals?%s", filter)
	request, _ := http.NewRequest("GET", urlWithFilter, nil)
	get, err := c.azureRequest(key, request)
	if err != nil {
		log.Println("Unable to get azure service principals.")
		return AzureServicePrincipals{}, err
	}

	var sps AzureServicePrincipals
	if err = json.NewDecoder(get.Body).Decode(&sps); err != nil {
		log.Println("Unable to decode azure web app response.")
		return AzureServicePrincipals{}, err
	}
	return sps, nil
}

func (c *AzureClient) GetAppRoleAssignedTo(key []byte, servicePrincipalId string) (AzureAppRoleAssignments, error) {

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals/%s/appRoleAssignedTo", servicePrincipalId)
	request, _ := http.NewRequest("GET", url, nil)
	get, err := c.azureRequest(key, request)
	if err != nil {
		log.Println("Unable to get azure app role assignments.")
		return AzureAppRoleAssignments{}, err
	}

	var assignments AzureAppRoleAssignments
	if err = json.NewDecoder(get.Body).Decode(&assignments); err != nil {
		log.Println("Unable to decode azure web app response.")
		return AzureAppRoleAssignments{}, err
	}
	return assignments, nil
}

func (c *AzureClient) DecodeKey(key []byte) (AzureKey, error) {
	var decoded AzureKey
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&decoded)
	return decoded, err
}

func (c *AzureClient) AccessTokenRequest(decoded AzureKey) (AzureAccessToken, error) {
	var accessToken AzureAccessToken
	tokenUrl := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", decoded.Tenant)
	postBody := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s&scope=https://graph.microsoft.com/.default", decoded.AppId, decoded.Secret)
	tokenResponse, tokenErr := c.HttpClient.Post(tokenUrl, "", strings.NewReader(postBody))
	if tokenErr != nil {
		return accessToken, tokenErr
	}
	err := json.NewDecoder(tokenResponse.Body).Decode(&accessToken)
	return accessToken, err
}

func (c *AzureClient) azureRequest(key []byte, request *http.Request) (*http.Response, error) {
	decoded, keyErr := c.DecodeKey(key)
	if keyErr != nil {
		log.Println("Unable to decode azure provider key.")
		return nil, keyErr
	}
	accessToken, tokenErr := c.AccessTokenRequest(decoded)
	if tokenErr != nil {
		log.Println("Unable to find azure web applications.")
		return nil, tokenErr
	}
	request.Header = http.Header{
		"ConsistencyLevel": []string{"eventual"},
		"Content-Type":     []string{"application/json"},
		"Authorization":    []string{fmt.Sprintf("Bearer %s", accessToken.Token)},
	}
	get, getError := c.HttpClient.Do(request)
	if getError != nil {
		log.Println("Unable to find azure web applications.")
		return nil, getError
	}
	log.Printf("Azure response %s.\n", get.Status)

	return get, nil
}
