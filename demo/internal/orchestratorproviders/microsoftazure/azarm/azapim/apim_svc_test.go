package azapim_test

import (
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armclientsupport"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/azuretestsupport/armtestsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewArmApimSvc_DefaultTransport(t *testing.T) {
	svc, err := azapim.NewArmApimSvc("", nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestNewArmApimSvc_OverrideTransport(t *testing.T) {
	h := &http.Client{}
	clientOpts := armclientsupport.NewArmClientOptions(h)
	svc, err := azapim.NewArmApimSvc("", azuretestsupport.AzureTokenCredential(h), clientOpts)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

// TestGetApimServiceInfo_NoMatchingService asserts no error when no matching
// apim service found with paging
func TestGetApimServiceInfo_NoMatchingService(t *testing.T) {
	apimServiceClient := apim_testsupport.NewMockApimServiceClient()
	opt := azapim.WithApimServiceClient(apimServiceClient)
	svc, _ := azapim.NewArmApimSvc("", nil, nil, opt)

	serviceUrl := armtestsupport.ApimServiceGatewayUrl

	apimServiceClient.ExpectNewListPager(
		apim_testsupport.ApimServiceListResponse("http://someurl1"),
		apim_testsupport.ApimServiceListResponse("http://someurl2"),
		apim_testsupport.ApimServiceListResponse("http://someurl3"))

	serviceInfo, err := svc.GetApimServiceInfo(serviceUrl)
	assert.NoError(t, err)
	assert.Empty(t, serviceInfo)
}

// TestGetApimServiceInfo_SinglePage asserts service info is returned
// when matching apim service found in a single page of results using a fake pager
func TestGetApimServiceInfo_SinglePage(t *testing.T) {
	apimServiceClient := apim_testsupport.NewMockApimServiceClient()
	opt := azapim.WithApimServiceClient(apimServiceClient)

	svc, _ := azapim.NewArmApimSvc("", nil, nil, opt)

	serviceUrl := armtestsupport.ApimServiceGatewayUrl
	apimServiceClient.ExpectNewListPager(apim_testsupport.ApimServiceListResponse(serviceUrl))

	serviceInfo, err := svc.GetApimServiceInfo(serviceUrl)
	assert.NoError(t, err)
	assert.Equal(t, armtestsupport.ApimServiceGatewayUrl, serviceInfo.ServiceUrl)
}

// TestGetApimServiceInfo_SinglePage asserts service info is returned
// when matching apim service found in a multiple pages of results using a fake pager
func TestGetApimServiceInfo_MultiplePages(t *testing.T) {
	apimServiceClient := apim_testsupport.NewMockApimServiceClient()
	opt := azapim.WithApimServiceClient(apimServiceClient)
	svc, _ := azapim.NewArmApimSvc("", nil, nil, opt)

	serviceUrl := armtestsupport.ApimServiceGatewayUrl

	apimServiceClient.ExpectNewListPager(
		apim_testsupport.ApimServiceListResponse("http://someurl1"),
		apim_testsupport.ApimServiceListResponse("http://someurl2"),
		apim_testsupport.ApimServiceListResponse(serviceUrl))

	serviceInfo, err := svc.GetApimServiceInfo(serviceUrl)
	assert.NoError(t, err)
	assert.Equal(t, armtestsupport.ApimServiceGatewayUrl, serviceInfo.ServiceUrl)
}

// TestAzureApimListService_ErrorResp asserts error is returned
// when azure api returns a 400
func TestListService_AzureApiBadRequest(t *testing.T) {
	httpClient := armtestsupport.FakeTokenCredentialHttpClient(armtestsupport.Issuer)
	service := buildApimSvc(httpClient)
	reqUrl := apim_testsupport.ListServiceUrl()
	httpClient.AddRequest("GET", reqUrl, http.StatusBadRequest, []byte(""))
	_, err := service.GetApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "400")
}

// TestListService_AzureApiOK asserts no error is returned
// when azure api returns a 200
func TestListService_AzureApiOK(t *testing.T) {
	httpClient := armtestsupport.FakeTokenCredentialHttpClient(armtestsupport.Issuer)
	service := buildApimSvc(httpClient)
	resp := apim_testsupport.ApimServiceListResponse(armtestsupport.ApimServiceGatewayUrl)
	theBytes, _ := json.Marshal(resp)
	reqUrl := apim_testsupport.ListServiceUrl()
	httpClient.AddRequest("GET", reqUrl, http.StatusOK, theBytes)

	_, err := service.GetApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)
	assert.NoError(t, err)
}

// BuildApimSvc - builds ArmApimSvc with a mock http client
func buildApimSvc(mockHttpClient *testsupport.MockHTTPClient) azapim.ArmApimSvc {
	key := azuretestsupport.AzureKeyBytes()
	factory, _ := azapim.NewSvcFactory(key, mockHttpClient)
	service, _ := factory.NewApimSvc()
	return service
}
