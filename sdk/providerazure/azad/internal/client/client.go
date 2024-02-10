package client

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azurecommon"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
)

type AzureGraphClient interface {
	Applications() *applications.ApplicationsRequestBuilder
}

type azureGraphClient struct {
	internal *msgraphsdk.GraphServiceClient
}

func NewAzureGraphClient(key []byte, httpClient azurecommon.HTTPClient) (AzureGraphClient, error) {
	scope := "https://graph.microsoft.com/.default"
	creds, err := azurecommon.ClientSecretCredentials(key, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure graph client. error: %w", err)
		//return nil, err
	}

	internal, err := msgraphsdk.NewGraphServiceClientWithCredentials(creds, []string{scope})
	if err != nil {
		return nil, err
	}

	return &azureGraphClient{internal: internal}, nil
}

func (g *azureGraphClient) Applications() *applications.ApplicationsRequestBuilder {
	return g.internal.Applications()
}
