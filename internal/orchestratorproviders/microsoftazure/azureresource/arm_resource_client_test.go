package azureresource_test

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/url"
	"testing"
)

type armResourceClient struct {
	mockClient *testsupport.MockHTTPClient
}

func newMockArnResourceClient() *armResourceClient {
	httpClient := armtestsupport.MockAuthorizedHttpClient(armtestsupport.Issuer)
	return &armResourceClient{mockClient: httpClient}
}

func listApimResourcesUrl() string {
	params := url.Values{
		"$filter":     {"resourceType eq 'Microsoft.ApiManagement/service'"},
		"api-version": {armtestsupport.ApiVersion},
	}

	resUrl := fmt.Sprintf("%s/%s/resources?%s",
		armtestsupport.AzureSubscriptionsBaseUrl, azuretestsupport.AzureSubscription, params.Encode())

	log.Println(resUrl)
	return resUrl
	//m.mockClient.AddRequest("GET", url, http.)
}

func (m *armResourceClient) expectListResources(expStatus int, expBody []byte) {
	reqUrl := listApimResourcesUrl()
	resId := "/subscriptions/f2f21609-3ca6-40dc-9a2d-511d705c49f5/resourceGroups/canarybankv2/providers/Microsoft.ApiManagement/service/canarybankapi"
	resName := "canarybankapi"
	resType := "Microsoft.ApiManagement/service"
	resource := armresources.GenericResourceExpanded{
		ID:   &resId,
		Name: &resName,
		Type: &resType,
	}

	resp := armresources.ClientListResponse{
		ResourceListResult: armresources.ResourceListResult{
			Value:    []*armresources.GenericResourceExpanded{&resource},
			NextLink: nil,
		},
	}

	log.Println("TestGetApiManagementResources Before marshalling ")
	theBytes, err := json.Marshal(resp)
	log.Println("TestGetApiManagementResources marshall error=", err)
	log.Println("TestGetApiManagementResources After marshalling ", string(theBytes))

	/*errCode := "some error code"
	errMsg := "some error msg"
	armresources.ClientListResponse{}
	armresources.ErrorResponse{
		Code:    &errCode,
		Message: &errMsg,
	}*/

	theBytes = []byte("")
	m.mockClient.AddRequest("GET", reqUrl, http.StatusBadRequest, theBytes)
}

func TestGetApiManagementResources(t *testing.T) {
	key := azuretestsupport.AzureClientKey()
	client := newMockArnResourceClient()
	factory, err := microsoftazure.NewApimProviderSvcFactory(key, client.mockClient)

	assert.NoError(t, err)

	client.expectListResources(http.StatusBadGateway, []byte(""))
	service, err := factory.NewArmResourceSvc()
	assert.NoError(t, err)
	resources, err := service.GetApiManagementResources()
	assert.Error(t, err)
	assert.Nil(t, resources)
}
