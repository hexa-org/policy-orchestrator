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

type azureKey struct {
	AppId        string `json:"appId"`
	Secret       string `json:"secret"`
	Tenant       string `json:"tenant"`
	Subscription string `json:"subscription"`
}

type azureAccessToken struct {
	Token string `json:"access_token"`
}

type azureWebApps struct {
	List []azureWebApp `json:"value"`
}

type azureWebApp struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

func (c *AzureClient) GetWebApplications(key []byte) (apps []provider.ApplicationInfo, err error) {

	var decoded azureKey
	err = json.NewDecoder(bytes.NewReader(key)).Decode(&decoded)
	if err != nil {
		log.Println("Unable to decode azure provider key.")
		return []provider.ApplicationInfo{}, err
	}

	tokenUrl := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", decoded.Tenant)
	postBody := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s&resource=https://management.azure.com/", decoded.AppId, decoded.Secret)
	tokenResponse, tokenErr := c.HttpClient.Post(tokenUrl, "", strings.NewReader(postBody))
	if tokenErr != nil {
		log.Println("Unable to find azure web applications.")
		return []provider.ApplicationInfo{}, tokenErr
	}

	var accessToken azureAccessToken
	err = json.NewDecoder(tokenResponse.Body).Decode(&accessToken)

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Web/sites?api-version=2021-02-01", decoded.Subscription)
	getRequest, _ := http.NewRequest("GET", url, nil)
	getRequest.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{fmt.Sprintf("Bearer %s", accessToken.Token)},
	}
	get, getError := c.HttpClient.Do(getRequest)
	if getError != nil {
		log.Println("Unable to find azure web applications.")
		return []provider.ApplicationInfo{}, getError
	}
	log.Printf("Azure response %s.\n", get.Status)

	var webapps azureWebApps
	if err = json.NewDecoder(get.Body).Decode(&webapps); err != nil {
		log.Println("Unable to decode azure web app response.")
		return []provider.ApplicationInfo{}, err
	}

	for _, app := range webapps.List {
		log.Printf("Found azure app service web app %s.\n", app.Name)
		apps = append(apps, provider.ApplicationInfo{ObjectID: app.ID, Name: app.Name, Description: app.Kind})
	}

	return apps, err
}
