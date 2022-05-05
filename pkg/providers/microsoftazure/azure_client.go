package microsoftazure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
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
	List []AzureAppRoleAssignment `json:"value"`
}

type AzureAppRoleAssignment struct {
	ID                   string `json:"id"`
	AppRoleId            string `json:"appRoleId"`
	PrincipalDisplayName string `json:"principalDisplayName"`
	PrincipalId          string `json:"principalId"`
	PrincipalType        string `json:"principalType"`
	ResourceDisplayName  string `json:"resourceDisplayName"`
	ResourceId           string `json:"resourceId"`
}

func (c *AzureClient) GetWebApplications(key []byte) ([]orchestrator.ApplicationInfo, error) {

	request, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/applications", nil)
	get, err := c.azureRequest(key, request)
	if err != nil {
		log.Println("Unable to get azure web applications.")
		return []orchestrator.ApplicationInfo{}, err
	}

	var webapps azureWebApps
	if err = json.NewDecoder(get.Body).Decode(&webapps); err != nil {
		log.Println("Unable to decode azure web app response.")
		return []orchestrator.ApplicationInfo{}, err
	}

	var apps []orchestrator.ApplicationInfo
	for _, app := range webapps.List {
		log.Printf("Found azure app service web app %s.\n", app.Name)
		apps = append(apps, orchestrator.ApplicationInfo{ObjectID: app.ID, Name: app.Name, Description: app.AppID})
	}
	return apps, err
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

func (c *AzureClient) SetAppRoleAssignedTo(key []byte, servicePrincipalId string, assignments []AzureAppRoleAssignment) error {
	existingRoleAssignments, err := c.GetAppRoleAssignedTo(key, servicePrincipalId)
	if err != nil {
		log.Println("Unable to get azure app role assignments.")
		return err
	}
	addErr := c.AddAppRolesAssignedTo(key, servicePrincipalId, c.ShouldAdd(assignments, existingRoleAssignments))
	if addErr != nil {
		log.Println("Unable to add azure app role assignments.")
		return addErr
	}
	removeErr := c.DeleteAppRolesAssignedTo(key, servicePrincipalId, c.ShouldRemove(existingRoleAssignments, assignments))
	if removeErr != nil {
		log.Println("Unable to delete azure app role assignments.")
		return removeErr
	}
	return nil
}

func (c *AzureClient) ShouldAdd(assignments []AzureAppRoleAssignment, existingRoleAssignments AzureAppRoleAssignments) []AzureAppRoleAssignment {
	var shouldAdd []AzureAppRoleAssignment
	for _, assignment := range assignments {
		var contains = false
		for _, existingAssignment := range existingRoleAssignments.List {
			if strings.Contains(assignment.PrincipalId, existingAssignment.PrincipalId) {
				contains = true
			}
		}
		if !contains {
			shouldAdd = append(shouldAdd, assignment)
		}
	}
	return shouldAdd
}

func (c *AzureClient) ShouldRemove(existingRoleAssignments AzureAppRoleAssignments, assignments []AzureAppRoleAssignment) []string {
	var shouldRemove []string
	for _, existingAssignment := range existingRoleAssignments.List {
		var contains = false
		for _, assignment := range assignments {
			if strings.Contains(assignment.PrincipalId, existingAssignment.PrincipalId) {
				contains = true
			}
		}
		if !contains {
			shouldRemove = append(shouldRemove, existingAssignment.ID)
		}
	}
	return shouldRemove
}

type azureAppRoleAssignmentPost struct {
	AppRoleId   string `json:"appRoleId"`
	PrincipalId string `json:"principalId"`
	ResourceId  string `json:"resourceId"`
}

func (c *AzureClient) AddAppRolesAssignedTo(key []byte, servicePrincipalId string, assignments []AzureAppRoleAssignment) (err error) {
	for _, assignment := range assignments {
		var buf bytes.Buffer
		ra := azureAppRoleAssignmentPost{assignment.AppRoleId, assignment.PrincipalId, assignment.ResourceId}
		_ = json.NewEncoder(&buf).Encode(ra)
		url := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals/%s/appRoleAssignedTo", servicePrincipalId)
		request, _ := http.NewRequest("POST", url, bytes.NewReader(buf.Bytes()))
		_, err = c.azureRequest(key, request)
		if err != nil {
			log.Println("Unable to add azure app role assignments.")
			return err
		}
	}
	return err
}

func (c *AzureClient) DeleteAppRolesAssignedTo(key []byte, servicePrincipalId string, assignmentIds []string) (err error) {
	for _, assignmentId := range assignmentIds {
		url := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals/%s/appRoleAssignedTo/%s", servicePrincipalId, assignmentId)
		request, _ := http.NewRequest("DELETE", url, nil)
		_, err = c.azureRequest(key, request)
		if err != nil {
			log.Println("Unable to delete azure app role assignments.")
			return err
		}
	}
	return err
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
	return c.HttpClient.Do(request)
}
