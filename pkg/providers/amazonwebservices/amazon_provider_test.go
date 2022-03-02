package amazonwebservices_test

import (
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/amazonwebservices"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/amazonwebservices/test"
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
	p := &amazonwebservices.AmazonProvider{Client: &cognitoidentityprovider.Client{}}
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
	info := provider.IntegrationInfo{Name: "amazon", Key: key}
	p := &amazonwebservices.AmazonProvider{Client: &cognitoidentityprovider.Client{}}
	_, err := p.DiscoverApplications(info)
	assert.Equal(t, "operation error Cognito Identity Provider: ListUserPools, expected endpoint resolver to not be nil", err.Error())
}

func TestAmazonProvider_DiscoverApplications_withOtherProvider(t *testing.T) {
	p := &amazonwebservices.AmazonProvider{}
	info := provider.IntegrationInfo{Name: "not_amazon", Key: []byte("aKey")}
	_, err := p.DiscoverApplications(info)
	assert.NoError(t, err)
}

func TestAmazonProvider_ListUserPools(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	p := &amazonwebservices.AmazonProvider{Client: mockClient}
	pools, _ := p.ListUserPools()
	assert.Equal(t, "anId", pools[0].ObjectID)
	assert.Equal(t, "aName", pools[0].Name)
}

func TestAmazonProvider_ListUserPools_withError(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	mockClient.Err = errors.New("oops")
	p := &amazonwebservices.AmazonProvider{Client: mockClient}
	_, err := p.ListUserPools()
	assert.Error(t, err)
}

func TestAmazonProvider_GetPolicyInfo(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	p := &amazonwebservices.AmazonProvider{Client: mockClient}
	info, _ := p.GetPolicyInfo(provider.IntegrationInfo{}, provider.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Equal(t, 1, len(info))
	assert.Equal(t, "aUser@amazon.com", info[0].Subject.AuthenticatedUsers[0])
	assert.Equal(t, "anObjectId", info[0].Object.Resources[0])
}

func TestAmazonProvider_GetPolicyInfo_withError(t *testing.T) {
	mockClient := &amazonwebservices_test.MockClient{}
	mockClient.Err = errors.New("oops")
	p := &amazonwebservices.AmazonProvider{Client: mockClient}
	_, err := p.GetPolicyInfo(provider.IntegrationInfo{}, provider.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Error(t, err)
}

func TestAmazonProvider_SetPolicyInfo(t *testing.T) {
	p := &amazonwebservices.AmazonProvider{}
	err := p.SetPolicyInfo(provider.IntegrationInfo{}, provider.ApplicationInfo{}, provider.PolicyInfo{})
	assert.NoError(t, err)
}
