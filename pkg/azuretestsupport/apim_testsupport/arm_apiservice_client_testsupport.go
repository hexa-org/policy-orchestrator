package apim_testsupport

import (
	"encoding/json"
	"fmt"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"net/http"
)

func ApimServiceId() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ApiManagement/%s",
		azuretestsupport.AzureSubscription,
		armtestsupport.ApimResourceGroupName,
		armtestsupport.ApimServiceName)
}

func ListServiceUrl() string {
	return fmt.Sprintf("%s/%s/providers/Microsoft.ApiManagement/service?api-version=2021-08-01",
		armtestsupport.AzureSubscriptionsBaseUrl, azuretestsupport.AzureSubscription)
}
func (m *AzureApimHttpClient) ExpectListService() {
	id := ApimServiceId()
	name := armtestsupport.ApimServiceName
	resType := "Microsoft.ApiManagement/service"
	gatewayUrl := armtestsupport.ApimServiceGatewayUrl
	props := azarmapim.ServiceProperties{GatewayURL: &gatewayUrl}
	resp := azarmapim.ServiceClientListResponse{
		ServiceListResult: azarmapim.ServiceListResult{
			Value: []*azarmapim.ServiceResource{{
				Properties: &props,
				ID:         &id,
				Name:       &name,
				Type:       &resType,
			}},
		},
	}

	theBytes, _ := json.Marshal(resp)
	reqUrl := ListServiceUrl()
	m.HttpClient.AddRequest("GET", reqUrl, http.StatusOK, theBytes)
}

func (m *AzureApimHttpClient) expectGetApiClient() {
	url := fmt.Sprintf("%s/%s/resourceGroups/%s/providers/Microsoft.ApiManagement/service/%s/apis/%s?api-version=2021-08-01",
		armtestsupport.AzureSubscriptionsBaseUrl, azuretestsupport.AzureSubscription,
		armtestsupport.ApimResourceGroupName,
		armtestsupport.ApimServiceName,
		armtestsupport.ApimAppId)

	output := azarmapim.APIClientGetResponse{
		APIContract: azarmapim.APIContract{},
		ETag:        nil,
	}
	resp, _ := json.Marshal(output)
	m.HttpClient.AddRequest(http.MethodGet, url, http.StatusOK, resp)
}
