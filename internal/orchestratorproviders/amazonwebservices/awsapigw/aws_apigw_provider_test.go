package awsapigw_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/awstestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAwsApiGatewayProvider_NoOverrides(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	assert.NotNil(t, p)
}

func TestNewAwsApiGatewayProvider_GetProviderService_WithDynamodbClient(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()

	assert.NotNil(t, p)
	_, err := p.GetPolicyInfo(awstestsupport.IntegrationInfo(), awstestsupport.AppInfo())
	assert.ErrorContains(t, err, "StatusCode: 400")
}

func TestAwsApiGatewayProvider_GetProviderService_WithCognitoClient(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	integration := awstestsupport.IntegrationInfo()
	apps, err := p.DiscoverApplications(integration)
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, apps)
}

func TestAwsApiGatewayProvider_GetProviderService_InvalidKey(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	integration := awstestsupport.IntegrationInfo()
	integration.Key = []byte("a")
	apps, err := p.DiscoverApplications(integration)
	assert.ErrorContains(t, err, "invalid character 'a'")
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProvider_DiscoverApplications_InvalidProviderName(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	integration := awstestsupport.IntegrationInfo()
	integration.Name = "invalid"
	apps, _ := p.DiscoverApplications(integration)
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProvider_DiscoverApplications_Error(t *testing.T) {
	cognitoClient := &mockCognitoClient{}
	cognitoClient.expectListUserPools(nil, errors.New("some error"))

	opt := awsapigw.WithCognitoClientOverride(cognitoClient)
	p := awsapigw.NewAwsApiGatewayProvider(opt)
	apps, err := p.DiscoverApplications(awstestsupport.IntegrationInfo())
	assert.Error(t, err)
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProvider_DiscoverApplications(t *testing.T) {
	cognitoClient := &mockCognitoClient{}
	expApps := []orchestrator.ApplicationInfo{awstestsupport.AppInfo()}
	cognitoClient.expectListUserPools(expApps, nil)

	opt := awsapigw.WithCognitoClientOverride(cognitoClient)
	p := awsapigw.NewAwsApiGatewayProvider(opt)
	apps, err := p.DiscoverApplications(awstestsupport.IntegrationInfo())
	assert.NoError(t, err)
	assert.Len(t, apps, len(expApps))
}

func TestAwsApiGatewayProvider_GetPolicyInfo_GetProviderServiceError(t *testing.T) {
	p := awsapigw.NewAwsApiGatewayProvider()
	integration := awstestsupport.IntegrationInfo()
	integration.Key = []byte("a")
	apps, err := p.GetPolicyInfo(integration, awstestsupport.AppInfo())
	assert.ErrorContains(t, err, "invalid character 'a'")
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProvider_GetPolicyInfo_GetResourceRolesError(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	policyStoreSvc.expectGetResourceRoles(nil, errors.New("some-error"))

	opt := awsapigw.WithPolicyStoreSvcOverride(policyStoreSvc)
	p := awsapigw.NewAwsApiGatewayProvider(opt)
	actPolicies, err := p.GetPolicyInfo(awstestsupport.IntegrationInfo(), awstestsupport.AppInfo())
	assert.ErrorContains(t, err, "some-error")
	assert.NotNil(t, actPolicies)
	assert.Equal(t, 0, len(actPolicies))
}

func TestAwsApiGatewayProvider_GetPolicyInfo_GetResourceRolesEmptyResponse(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	policyStoreSvc.expectGetResourceRoles(nil, nil)

	opt := awsapigw.WithPolicyStoreSvcOverride(policyStoreSvc)
	p := awsapigw.NewAwsApiGatewayProvider(opt)
	actPolicies, err := p.GetPolicyInfo(awstestsupport.IntegrationInfo(), awstestsupport.AppInfo())
	assert.NoError(t, err)
	assert.NotNil(t, actPolicies)
	assert.Equal(t, 0, len(actPolicies))
}

func TestAwsApiGatewayProvider_GetPolicyInfo(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)

	opt := awsapigw.WithPolicyStoreSvcOverride(policyStoreSvc)
	p := awsapigw.NewAwsApiGatewayProvider(opt)
	actPolicies, err := p.GetPolicyInfo(awstestsupport.IntegrationInfo(), awstestsupport.AppInfo())
	assert.NoError(t, err)
	assert.NotNil(t, actPolicies)
	assert.Equal(t, len(existingActionRoles), len(actPolicies))
}
