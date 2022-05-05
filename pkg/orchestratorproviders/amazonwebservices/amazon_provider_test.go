package amazonwebservices_test

import (
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/amazonwebservices"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/amazonwebservices/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAmazonProvider_Credentials(t *testing.T) {
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "aRegion"
}
`)
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: &cognitoidentityprovider.Client{}}
	c := p.Credentials(key)
	assert.Equal(t, "anAccessKeyID", c.AccessKeyID)
}

func TestAmazonProvider_DiscoverApplications(t *testing.T) {
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "aRegion"
}
`)
	info := orchestrator.IntegrationInfo{Name: "amazon", Key: key}
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: &cognitoidentityprovider.Client{}}
	_, err := p.DiscoverApplications(info)
	assert.Equal(t, "operation error Cognito Identity Provider: ListUserPools, expected endpoint resolver to not be nil", err.Error())
}

func TestAmazonProvider_DiscoverApplications_withOtherProvider(t *testing.T) {
	p := &amazonwebservices.AmazonProvider{}
	info := orchestrator.IntegrationInfo{Name: "not_amazon", Key: []byte("aKey")}
	_, err := p.DiscoverApplications(info)
	assert.NoError(t, err)
	assert.Nil(t, p.CognitoClientOverride)
}

func TestAmazonProvider_ListUserPools(t *testing.T) {
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "aRegion"
}
`)
	mockClient := &amazonwebservices_test.MockClient{}
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	info := orchestrator.IntegrationInfo{Name: "amazon", Key: key}
	pools, _ := p.ListUserPools(info)
	assert.Equal(t, "anId", pools[0].ObjectID)
	assert.Equal(t, "aName", pools[0].Name)
}

func TestAmazonProvider_ListUserPools_withError(t *testing.T) {
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "aRegion"
}
`)
	mockClient := &amazonwebservices_test.MockClient{Errs: map[string]error{}}
	mockClient.Errs["ListUserPools"] = errors.New("oops")
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	info := orchestrator.IntegrationInfo{Name: "amazon", Key: key}
	_, err := p.ListUserPools(info)
	assert.Error(t, err)
}

func TestAmazonProvider_GetPolicyInfo(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	info, _ := p.GetPolicyInfo(orchestrator.IntegrationInfo{}, orchestrator.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Equal(t, 1, len(info))
	assert.Equal(t, "aUser:aUser@amazon.com", info[0].Subject.AuthenticatedUsers[0])
	assert.Equal(t, "anObjectId", info[0].Object.Resources[0])
}

func TestAmazonProvider_GetPolicyInfo_withError(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{Errs: map[string]error{}}
	mockClient.Errs["ListUsers"] = errors.New("oops")
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	_, err := p.GetPolicyInfo(orchestrator.IntegrationInfo{}, orchestrator.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Error(t, err)
}

func TestAmazonProvider_ShouldEnable(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	shouldAdd := p.ShouldEnable([]string{"aUser@amazon.com", "yetAnotherUser@amazon.com"}, []string{"anotherUser@amazon.com"})
	assert.Equal(t, []string{"anotherUser@amazon.com"}, shouldAdd)
}

func TestAmazonProvider_ShouldDisable(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	shouldAdd := p.ShouldDisable([]string{"aUser@amazon.com", "yetAnotherUser@amazon.com"}, []string{"yetAnotherUser@amazon.com"})
	assert.Equal(t, []string{"aUser@amazon.com"}, shouldAdd)
}

func TestAmazonProvider_SetPolicyInfo(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	err := p.SetPolicyInfo(orchestrator.IntegrationInfo{}, orchestrator.ApplicationInfo{}, []policysupport.PolicyInfo{{
		Subject: policysupport.SubjectInfo{AuthenticatedUsers: []string{"aUser:aUser@amazon.com", "anotherUser:anotherUser@amazon.com"}},
		Object:  policysupport.ObjectInfo{Resources: []string{"aResource"}},
	}})
	assert.NoError(t, err)
}

func TestAmazonProvider_SetPolicyInfo_withListErr(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{Errs: map[string]error{}}
	mockClient.Errs["ListUsers"] = errors.New("oops")
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	err := p.SetPolicyInfo(orchestrator.IntegrationInfo{}, orchestrator.ApplicationInfo{}, []policysupport.PolicyInfo{{}})
	assert.Error(t, err)
}

func TestAmazonProvider_SetPolicyInfo_withEnableErr(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{Errs: map[string]error{}}
	mockClient.Errs["AdminEnableUser"] = errors.New("oops")
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	err := p.SetPolicyInfo(orchestrator.IntegrationInfo{}, orchestrator.ApplicationInfo{}, []policysupport.PolicyInfo{{
		Subject: policysupport.SubjectInfo{AuthenticatedUsers: []string{"aUser:aUser@amazon.com", "anotherUser:anotherUser@amazon.com"}},
		Object:  policysupport.ObjectInfo{Resources: []string{"aResource"}},
	}})
	assert.Error(t, err)
}

func TestAmazonProvider_SetPolicyInfo_withDisableErr(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{Errs: map[string]error{}}
	mockClient.Errs["AdminDisableUser"] = errors.New("oops")
	p := &amazonwebservices.AmazonProvider{CognitoClientOverride: mockClient}
	err := p.SetPolicyInfo(orchestrator.IntegrationInfo{}, orchestrator.ApplicationInfo{}, []policysupport.PolicyInfo{{}})
	assert.Error(t, err)
}
