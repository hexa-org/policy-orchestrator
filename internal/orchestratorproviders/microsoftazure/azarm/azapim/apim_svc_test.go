package azapim_test

import (
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewArmApimSvc(t *testing.T) {
	svc, err := azapim.NewArmApimSvc("", nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestGetApimServiceInfo_NoMatchingService(t *testing.T) {
	apimServiceClient := apim_testsupport.NewMockApimServiceClient()
	opt := azapim.WithApimServiceClient(apimServiceClient)
	svc, _ := azapim.NewArmApimSvc("", nil, nil, opt)

	serviceUrl := armtestsupport.ApimServiceGatewayUrl

	apimServiceClient.ExpectNewListPager(
		oneApimServicePage("http://someurl1"),
		oneApimServicePage("http://someurl2"),
		oneApimServicePage("http://someurl3"))

	serviceInfo, err := svc.GetApimServiceInfo(serviceUrl)
	assert.NoError(t, err)
	assert.Empty(t, serviceInfo)
}

func TestGetApimServiceInfo_SinglePage(t *testing.T) {
	apimServiceClient := apim_testsupport.NewMockApimServiceClient()
	opt := azapim.WithApimServiceClient(apimServiceClient)

	svc, _ := azapim.NewArmApimSvc("", nil, nil, opt)

	serviceUrl := armtestsupport.ApimServiceGatewayUrl
	apimServiceClient.ExpectNewListPager(oneApimServicePage(serviceUrl))

	serviceInfo, err := svc.GetApimServiceInfo(serviceUrl)
	assert.NoError(t, err)
	assert.Equal(t, armtestsupport.ApimServiceGatewayUrl, serviceInfo.ServiceUrl)
}

func TestGetApimServiceInfo_MultiplePages(t *testing.T) {
	apimServiceClient := apim_testsupport.NewMockApimServiceClient()
	opt := azapim.WithApimServiceClient(apimServiceClient)
	svc, _ := azapim.NewArmApimSvc("", nil, nil, opt)

	serviceUrl := armtestsupport.ApimServiceGatewayUrl

	apimServiceClient.ExpectNewListPager(
		oneApimServicePage("http://someurl1"),
		oneApimServicePage("http://someurl2"),
		oneApimServicePage(serviceUrl))

	serviceInfo, err := svc.GetApimServiceInfo(serviceUrl)
	assert.NoError(t, err)
	assert.Equal(t, armtestsupport.ApimServiceGatewayUrl, serviceInfo.ServiceUrl)
}

func TestAzureApimListService_ErrorResp(t *testing.T) {
	mockApiClient := apim_testsupport.MockApimHttpClient()
	service := apim_testsupport.BuildApimSvc(mockApiClient.HttpClient)
	reqUrl := apim_testsupport.ListServiceUrl()
	mockApiClient.HttpClient.AddRequest("GET", reqUrl, http.StatusBadRequest, []byte(""))
	_, err := service.GetApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "400")
}

func TestAzureApimListService(t *testing.T) {
	mockApiClient := apim_testsupport.MockApimHttpClient()
	service := apim_testsupport.BuildApimSvc(mockApiClient.HttpClient)
	mockApiClient.ExpectListService()
	_, err := service.GetApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)
	assert.NoError(t, err)
}

func TestGetResourceRoles_NoServiceInfo(t *testing.T) {
	svc, _ := azapim.NewArmApimSvc("", nil, nil)
	roles, err := svc.GetResourceRoles(armmodel.ApimServiceInfo{})
	assert.NoError(t, err)
	assert.NotNil(t, roles)
	assert.Equal(t, []providerscommon.ResourceActionRoles{}, roles)
}

func oneApimServicePage(gatewayUrl string) azarmapim.ServiceClientListResponse {
	id := apim_testsupport.ApimServiceId()
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
