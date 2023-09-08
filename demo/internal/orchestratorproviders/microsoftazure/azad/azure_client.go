package azad

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azurecommon"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type AzureClient interface {
	GetAzureApplications(key []byte) ([]AzureWebApp, error)
	GetWebApplications(key []byte) ([]orchestrator.ApplicationInfo, error)
	GetServicePrincipals(key []byte, appId string) (AzureServicePrincipals, error)
	GetUserInfoFromPrincipalId(key []byte, principalId string) (AzureUser, error)
	GetPrincipalIdFromEmail(key []byte, email string) (string, error)
	GetAppRoleAssignedTo(key []byte, servicePrincipalId string) (AzureAppRoleAssignments, error)
	SetAppRoleAssignedTo(key []byte, servicePrincipalId string, assignments []AzureAppRoleAssignment) error
}

type azureClient struct {
	HttpClient azurecommon.HTTPClient
}

type AzureAccessToken struct {
	Token string `json:"access_token"`
}

type azureWebApps struct {
	List []AzureWebApp `json:"value"`
}

type AzureWebApp struct {
	ID             string       `json:"id"`
	AppID          string       `json:"appId"`
	Name           string       `json:"displayName"`
	IdentifierUris []string     `json:"identifierUris"`
	Web            azureWebInfo `json:"web"`
}

type azureWebInfo struct {
	HomePageUrl string `json:"homePageUrl"`
}
type AzureServicePrincipals struct {
	List []azureServicePrincipal `json:"value"`
}

type azureServicePrincipal struct {
	ID       string         `json:"id"`
	Name     string         `json:"displayName"`
	AppRoles []AzureAppRole `json:"appRoles"`
}

type AzureAppRole struct {
	AllowedMemberTypes []string `json:"allowedMemberTypes"`
	Description        string   `json:"description"`
	DisplayName        string   `json:"displayName"`
	ID                 string   `json:"id"`
	IsEnabled          bool     `json:"isEnabled"`
	Origin             string   `json:"origin"`
	Value              string   `json:"value"`
}

type AzureAppRoleAssignments struct {
	List []AzureAppRoleAssignment `json:"value"`
}

type AzureUser struct {
	PrincipalId string `json:"id"`
	Name        string `json:"userPrincipalName"`
	Email       string `json:"mail"`
}

type AzureUsers struct {
	List []AzureUser `json:"value"`
}

type AzureAppRoleAssignment struct {
	ID                   string `json:"id"`
	AppRoleId            string `json:"appRoleId" validate:"required"`
	PrincipalDisplayName string `json:"principalDisplayName"`
	PrincipalId          string `json:"principalId"`
	PrincipalType        string `json:"principalType"`
	ResourceDisplayName  string `json:"resourceDisplayName"`
	ResourceId           string `json:"resourceId" validate:"required"`
}

func NewAzureClient(httpClient azurecommon.HTTPClient) AzureClient {
	if httpClient == nil {
		return &azureClient{HttpClient: &http.Client{}}
	}
	return &azureClient{HttpClient: httpClient}
}

func (c *azureClient) GetAzureApplications(key []byte) ([]AzureWebApp, error) {
	request, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/applications", nil)
	get, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
	if err != nil {
		log.Println("Unable to get azure web applications. Error=" + err.Error())
		return nil, err
	}

	if get.StatusCode != http.StatusOK {
		errMsg := "unable to get azure web applications. Unexpected status " + get.Status
		log.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	var webapps azureWebApps
	if err = json.NewDecoder(get.Body).Decode(&webapps); err != nil {
		log.Println("Unable to decode azure web app response.")
		return nil, err
	}

	return webapps.List, nil
}

func (c *azureClient) GetWebApplications(key []byte) ([]orchestrator.ApplicationInfo, error) {
	request, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/applications", nil)
	get, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
	if err != nil {
		log.Println("Unable to get azure web applications. Error=" + err.Error())
		return []orchestrator.ApplicationInfo{}, err
	}

	if get.StatusCode != http.StatusOK {
		errMsg := "unable to get azure web applications. Unexpected status " + get.Status
		log.Println(errMsg)
		return []orchestrator.ApplicationInfo{}, errors.New(errMsg)
	}

	var webapps azureWebApps
	if err = json.NewDecoder(get.Body).Decode(&webapps); err != nil {
		log.Println("Unable to decode azure web app response.")
		return []orchestrator.ApplicationInfo{}, err
	}

	var apps []orchestrator.ApplicationInfo
	for _, app := range webapps.List {
		log.Printf("Found azure app service web app %s.\n", app.Name)
		if app.Web.HomePageUrl != "" { // todo - a better way to find enterprise apps, WindowsAzureActiveDirectoryIntegratedApp?
			apps = append(apps, orchestrator.ApplicationInfo{
				ObjectID:    app.ID,
				Name:        app.Name,
				Description: app.AppID,
				Service:     "App Service",
			})
		}
	}
	return apps, err
}

func (c *azureClient) GetServicePrincipals(key []byte, appId string) (AzureServicePrincipals, error) {
	filter := fmt.Sprintf("$search=\"appId:%s\"", appId)
	urlWithFilter := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals?%s", filter)
	request, _ := http.NewRequest("GET", urlWithFilter, nil)
	get, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
	if err != nil {
		log.Println("Unable to get azure service principals. Error=" + err.Error())
		return AzureServicePrincipals{}, err
	}

	if get.StatusCode != http.StatusOK {
		errMsg := "unable to get azure service principals. Unexpected status " + get.Status
		log.Println(errMsg)
		return AzureServicePrincipals{}, errors.New(errMsg)
	}

	var sps AzureServicePrincipals
	if err = json.NewDecoder(get.Body).Decode(&sps); err != nil {
		log.Println("Unable to decode azure service principals response. Error=", err)
		return AzureServicePrincipals{}, err
	}
	return sps, nil
}

func (c *azureClient) GetUserInfoFromPrincipalId(key []byte, principalId string) (AzureUser, error) {
	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s", principalId)
	request, _ := http.NewRequest("GET", endpoint, nil)
	get, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
	if err != nil {
		log.Println("Unable to get azure user. Error=" + err.Error())
		return AzureUser{}, err
	}

	if get.StatusCode != http.StatusOK {
		errMsg := "unable to get azure user. Unexpected status " + get.Status
		log.Println(errMsg)
		return AzureUser{}, errors.New(errMsg)
	}

	var user AzureUser
	if err = json.NewDecoder(get.Body).Decode(&user); err != nil {
		log.Println("Unable to decode azure web app response. Error=" + err.Error())
		return AzureUser{}, err
	}

	return user, nil
}

func (c *azureClient) GetPrincipalIdFromEmail(key []byte, email string) (string, error) {
	query := fmt.Sprintf("https://graph.microsoft.com/v1.0/users?$select=id,mail&$filter=mail%%20eq%%20%%27%s%%27", url.QueryEscape(email))
	request, _ := http.NewRequest("GET", query, nil)
	get, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
	if err != nil {
		log.Println("Unable to get id for azure user. Error=" + err.Error())
		return "", err
	}

	if get.StatusCode != http.StatusOK {
		errMsg := "unable to get id for azure user. Unexpected status " + get.Status
		log.Println(errMsg)
		return "", errors.New(errMsg)
	}

	var userValues AzureUsers
	if err = json.NewDecoder(get.Body).Decode(&userValues); err != nil {
		log.Println("Unable to decode azure web app response. Error=", err)
		return "", err
	}
	return userValues.List[0].PrincipalId, nil
}

func (c *azureClient) GetAppRoleAssignedTo(key []byte, servicePrincipalId string) (AzureAppRoleAssignments, error) {
	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals/%s/appRoleAssignedTo", servicePrincipalId)
	request, _ := http.NewRequest("GET", endpoint, nil)
	get, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
	if err != nil {
		log.Println("Unable to get azure app role assignments. Error=" + err.Error())
		return AzureAppRoleAssignments{}, err
	}

	if get.StatusCode != http.StatusOK {
		errMsg := "unable to get id for azure app role assignments. Unexpected status " + get.Status
		log.Println(errMsg)
		return AzureAppRoleAssignments{}, errors.New(errMsg)
	}

	var assignments AzureAppRoleAssignments
	if err = json.NewDecoder(get.Body).Decode(&assignments); err != nil {
		log.Println("Unable to decode azure web app response. Error=" + err.Error())
		return AzureAppRoleAssignments{}, err
	}

	if assignments.List == nil {
		assignments.List = []AzureAppRoleAssignment{}
	}

	return assignments, nil
}

func (c *azureClient) SetAppRoleAssignedTo(key []byte, servicePrincipalId string, assignments []AzureAppRoleAssignment) error {
	validate := validator.New()
	vErr := validate.Var(assignments, "omitempty,dive")
	if vErr != nil {
		log.Println("Validate error ", vErr)
		return vErr
	}
	existingRoleAssignments, err := c.GetAppRoleAssignedTo(key, servicePrincipalId)
	if err != nil {
		log.Println("Unable to get azure app role assignments. Error=" + err.Error())
		return err
	}
	addErr := c.addAppRolesAssignedTo(key, servicePrincipalId, c.shouldAdd(assignments, existingRoleAssignments))
	if addErr != nil {
		log.Println("Unable to add azure app role assignments.")
		return addErr
	}
	removeErr := c.deleteAppRolesAssignedTo(key, servicePrincipalId, c.shouldRemove(existingRoleAssignments, assignments))
	if removeErr != nil {
		log.Println("Unable to delete azure app role assignments.")
		return removeErr
	}
	return nil
}

func (c *azureClient) shouldAdd(assignments []AzureAppRoleAssignment, existingRoleAssignments AzureAppRoleAssignments) []AzureAppRoleAssignment {
	var shouldAdd []AzureAppRoleAssignment
	for _, assignment := range assignments {
		if assignment.PrincipalId == "" {
			continue
		}

		exists := false
		for _, existingAssignment := range existingRoleAssignments.List {
			if existingAssignment.AppRoleId == assignment.AppRoleId &&
				existingAssignment.ResourceId == assignment.ResourceId &&
				existingAssignment.PrincipalId == assignment.PrincipalId {
				exists = true
				break
			}
		}

		if !exists {
			shouldAdd = append(shouldAdd, assignment)
		}
	}

	return shouldAdd
}

func (c *azureClient) shouldRemove(existingRoleAssignments AzureAppRoleAssignments, assignments []AzureAppRoleAssignment) []string {
	var shouldRemove []string

	for _, eAra := range existingRoleAssignments.List {
		doRemove := false
		for _, ara := range assignments {
			if eAra.AppRoleId == ara.AppRoleId && eAra.ResourceId == ara.ResourceId {
				if eAra.PrincipalId == ara.PrincipalId {
					doRemove = false
					break
				}
				doRemove = true
			}
		}

		if doRemove {
			shouldRemove = append(shouldRemove, eAra.ID)
		}
	}

	return shouldRemove
}

type azureAppRoleAssignmentPost struct {
	AppRoleId   string `json:"appRoleId"`
	PrincipalId string `json:"principalId"`
	ResourceId  string `json:"resourceId"`
}

func (c *azureClient) addAppRolesAssignedTo(key []byte, servicePrincipalId string, assignments []AzureAppRoleAssignment) (err error) {
	for _, assignment := range assignments {
		var buf bytes.Buffer
		ra := azureAppRoleAssignmentPost{assignment.AppRoleId, assignment.PrincipalId, servicePrincipalId} // the resource id is the service principal
		_ = json.NewEncoder(&buf).Encode(ra)
		endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals/%s/appRoleAssignedTo", servicePrincipalId)
		request, _ := http.NewRequest("POST", endpoint, bytes.NewReader(buf.Bytes()))
		response, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
		if err != nil {
			log.Println("Unable to add azure app role assignments. Error=", err)
			return err
		}

		if response.StatusCode != http.StatusCreated {
			errMsg := fmt.Sprintf("unable to add azure app role assignments. Unexpected status %d", response.StatusCode)
			log.Println(errMsg)
			return errors.New(errMsg)
		}
	}
	return err
}

func (c *azureClient) deleteAppRolesAssignedTo(key []byte, servicePrincipalId string, assignmentIds []string) (err error) {
	for _, assignmentId := range assignmentIds {
		endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/servicePrincipals/%s/appRoleAssignedTo/%s", servicePrincipalId, assignmentId)
		request, _ := http.NewRequest("DELETE", endpoint, nil)
		response, err := c.azureRequest(key, request, "https://graph.microsoft.com/.default")
		if err != nil {
			log.Println("Unable to delete azure app role assignments. Error=", err)
			return err
		}

		if response.StatusCode != http.StatusNoContent {
			errMsg := fmt.Sprintf("unable to delete azure app role assignments. Unexpected status %d", response.StatusCode)
			log.Println(errMsg)
			return errors.New(errMsg)
		}
	}
	return err
}

func (c *azureClient) decodeKey(key []byte) (azurecommon.AzureKey, error) {
	var decoded azurecommon.AzureKey
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&decoded)
	return decoded, err
}

func (c *azureClient) accessTokenRequest(decoded azurecommon.AzureKey, scope string) (AzureAccessToken, error) {
	var accessToken AzureAccessToken
	tokenUrl := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", decoded.Tenant)
	postBody := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s&scope=%s", decoded.AppId, decoded.Secret, scope)
	tokenResponse, tokenErr := c.HttpClient.Post(tokenUrl, "", strings.NewReader(postBody))
	if tokenErr != nil {
		return accessToken, tokenErr
	}

	err := json.NewDecoder(tokenResponse.Body).Decode(&accessToken)
	return accessToken, err
}

func (c *azureClient) azureRequest(key []byte, request *http.Request, scope string) (*http.Response, error) {
	decoded, keyErr := c.decodeKey(key)
	if keyErr != nil {
		log.Println("Unable to decode azure provider key. Error=", keyErr)
		return nil, keyErr
	}
	accessToken, tokenErr := c.accessTokenRequest(decoded, scope)
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
