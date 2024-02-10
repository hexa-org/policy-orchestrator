package apim_testsupport

import (
	"fmt"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/azuretestsupport/armtestsupport"
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

func ApimServiceListResponse(gatewayUrl string) azarmapim.ServiceClientListResponse {
	id := ApimServiceId()
	name := armtestsupport.ApimServiceName
	resType := "Microsoft.ApiManagement/service"

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
	return resp
}

func ApimServiceInfo(serviceUrl string) armmodel.ApimServiceInfo {
	return armmodel.ApimServiceInfo{
		ArmResource: armmodel.ArmResource{
			FullyQualifiedId: ApimServiceId(),
			ResourceGroup:    armtestsupport.ApimResourceGroupName,
			Type:             "Microsoft.ApiManagement/service",
			Name:             armtestsupport.ApimServiceName,
			DisplayName:      armtestsupport.ApimServiceName,
		},
		ServiceUrl: serviceUrl,
	}
}

/*
func (m *FakeApimHttpClient) expectGetApiClient() {
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
*/
