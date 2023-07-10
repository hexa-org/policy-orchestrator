package apimservice_test

import (
	"context"
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armclientsupport"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimservice"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azurecommon"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient_List(t *testing.T) {
	resp := apim_testsupport.ApimServiceListResponse(armtestsupport.ApimServiceGatewayUrl)
	theBytes, _ := json.Marshal(resp)
	reqUrl := apim_testsupport.ListServiceUrl()
	httpClient := armtestsupport.FakeTokenCredentialHttpClient(armtestsupport.Issuer)
	httpClient.AddRequest("GET", reqUrl, http.StatusOK, theBytes)

	client := apimServiceClient(httpClient)
	pager := client.NewListPager(nil)
	assert.True(t, pager.More())
	assert.NotNil(t, pager)

	page, err := pager.NextPage(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.False(t, pager.More())
}

func apimServiceClient(httpClient azurecommon.HTTPClient) apimservice.Client {
	tokenCredential, _ := azurecommon.ClientSecretCredentials(azuretestsupport.AzureKey(), httpClient)
	clientOptions := armclientsupport.NewArmClientOptions(httpClient)
	serviceClient := apimservice.NewClient(azuretestsupport.AzureSubscription, tokenCredential, clientOptions)
	return serviceClient
}
