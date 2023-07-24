package awscognito_test

import (
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscognito"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/cognitotestsupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestListUserPools_Error(t *testing.T) {
	mockHttpClient := cognitotestsupport.NewMockCognitoHTTPClient()
	mockHttpClient.MockListUserPoolsWithHttpStatus(http.StatusBadRequest)

	client := cognitoClient(mockHttpClient)
	pools, err := client.ListUserPools()
	assert.ErrorContains(t, err, "error StatusCode: 400")
	assert.ErrorContains(t, err, "ListUserPools")
	assert.Nil(t, pools)
}

func TestListUserPools_ResourceServersError(t *testing.T) {
	mockHttpClient := cognitotestsupport.NewMockCognitoHTTPClient()
	mockHttpClient.MockListUserPools()
	mockHttpClient.MockListResourceServersWithHttpStatus(http.StatusBadRequest, cognitoidentityprovider.ListResourceServersOutput{})

	client := cognitoClient(mockHttpClient)
	pools, err := client.ListUserPools()
	assert.ErrorContains(t, err, "error StatusCode: 400")
	assert.ErrorContains(t, err, "ListResourceServers")
	assert.Nil(t, pools)
}

func TestListUserPools_NoResourceServers(t *testing.T) {
	mockHttpClient := cognitotestsupport.NewMockCognitoHTTPClient()
	mockHttpClient.MockListUserPools()
	mockHttpClient.MockListResourceServers(cognitoidentityprovider.ListResourceServersOutput{})

	client := cognitoClient(mockHttpClient)
	pools, err := client.ListUserPools()
	assert.NoError(t, err)
	assert.Empty(t, pools)
}

func TestListUserPools_Success(t *testing.T) {
	mockHttpClient := cognitotestsupport.NewMockCognitoHTTPClient()
	mockHttpClient.MockListUserPools()
	mockHttpClient.MockListResourceServers(cognitotestsupport.WithResourceServer())

	client := cognitoClient(mockHttpClient)
	pools, err := client.ListUserPools()
	assert.NoError(t, err)
	assert.NotNil(t, pools)
	assert.Len(t, pools, 1)
	assert.Equal(t, cognitotestsupport.TestUserPoolId, pools[0].ObjectID)
	assert.Equal(t, cognitotestsupport.TestResourceServerName, pools[0].Name)
	assert.Equal(t, cognitotestsupport.TestResourceServerIdentifier, pools[0].Service)
	assert.Equal(t, "Cognito", pools[0].Description)
	assert.True(t, mockHttpClient.VerifyCalled())

}

func cognitoClient(mockHttpClient *cognitotestsupport.MockCognitoHTTPClient) awscognito.CognitoClient {
	info := cognitotestsupport.IntegrationInfo()
	client, _ := awscognito.NewCognitoClient(info.Key, awscommon.AWSClientOptions{
		HTTPClient:   mockHttpClient,
		DisableRetry: true})
	return client
}
