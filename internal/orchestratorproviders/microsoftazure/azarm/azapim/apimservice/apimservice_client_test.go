package apimservice_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestListService_ErrorResp(t *testing.T) {
	mockApiClient := apim_testsupport.MockApimHttpClient()
	service := apim_testsupport.BuildApimSvc(mockApiClient.HttpClient)
	reqUrl := apim_testsupport.ListServiceUrl()
	mockApiClient.HttpClient.AddRequest("GET", reqUrl, http.StatusBadRequest, []byte(""))
	_, err := service.GetApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "400")
}

func TestListService(t *testing.T) {
	mockApiClient := apim_testsupport.MockApimHttpClient()
	service := apim_testsupport.BuildApimSvc(mockApiClient.HttpClient)
	mockApiClient.ExpectListService()
	_, err := service.GetApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)
	assert.NoError(t, err)
}
