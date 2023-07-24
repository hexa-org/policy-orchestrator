package awsapigw_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/cognitotestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAwsApiGatewayProvider(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	assert.NotNil(t, p)
}

func TestAwsApiGatewayProvider_DiscoverApplications_InvalidProviderName(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	integration := cognitotestsupport.IntegrationInfo()
	integration.Name = "invalid"
	apps, _ := p.DiscoverApplications(integration)
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProvider_DiscoverApplications_CognitoClientError(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	integration := cognitotestsupport.IntegrationInfo()
	integration.Key = []byte("a")
	apps, err := p.DiscoverApplications(integration)
	assert.ErrorContains(t, err, "invalid character 'a'")
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProvider_DiscoverApplications_Error(t *testing.T) {
	cognitoClient := &mockCognitoClient{}
	cognitoClient.expectListUserPools(nil, errors.New("some error"))

	opt := awsapigw.WithCognitoClientOverride(cognitoClient)
	p := awsapigw.NewAwsApiGatewayProvider(opt)
	apps, err := p.DiscoverApplications(cognitotestsupport.IntegrationInfo())
	assert.Error(t, err)
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProvider_DiscoverApplications(t *testing.T) {
	cognitoClient := &mockCognitoClient{}
	expApps := []orchestrator.ApplicationInfo{cognitotestsupport.AppInfo()}
	cognitoClient.expectListUserPools(expApps, nil)

	opt := awsapigw.WithCognitoClientOverride(cognitoClient)
	p := awsapigw.NewAwsApiGatewayProvider(opt)
	apps, err := p.DiscoverApplications(cognitotestsupport.IntegrationInfo())
	assert.NoError(t, err)
	assert.Len(t, apps, len(expApps))
}
