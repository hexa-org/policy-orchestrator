package dynamodbpolicystore_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/table"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPolicyStoreSvc_Error(t *testing.T) {
	tableInfo, err := table.NewSimpleTableInfo(testhelper.TableName, testhelper.SimpleDynamodbItem{})
	svc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, []byte("$$"))
	assert.ErrorContains(t, err, "invalid character '$'")
	assert.Nil(t, svc)
}

func TestNewPolicyStoreSvc(t *testing.T) {
	tableInfo, err := table.NewSimpleTableInfo(testhelper.TableName, testhelper.SimpleDynamodbItem{})
	svc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, testhelper.AwsCredentialsForTest())
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestGetPolicies_ScanError(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	app := new(idp.AppInfo)
	c.ExpectScan(errors.New("some-error"))
	policies, err := svc.GetPolicies(*app)
	assert.ErrorContains(t, err, "some-error")
	assert.Nil(t, policies)
}

func TestGetPolicies_EmptyItemsFromScan(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	app := new(idp.AppInfo)
	c.ExpectScan(nil)
	policies, err := svc.GetPolicies(*app)
	assert.NoError(t, err)
	assert.NotNil(t, policies)
	assert.Empty(t, policies)
}
func TestGetPolicies(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	app := new(idp.AppInfo)
	expRar := testhelper.MakeResourceActionRoles()

	c.ExpectScan(nil, expRar)
	policies, err := svc.GetPolicies(*app)
	assert.NoError(t, err)
	assert.NotNil(t, policies)
	assert.Equal(t, []rar.ResourceActionRoles{expRar}, policies)
}

func TestSetPolicy(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	expRar := testhelper.MakeResourceActionRoles()
	c.ExpectUpdateItem(expRar, nil)
	err := svc.SetPolicy(expRar)
	assert.NoError(t, err)
}

func TestWithDynamicItemJson(t *testing.T) {
	svc, mockClient := newSvcWithDynamicTableDef()
	app := new(idp.AppInfo)
	expRar := testhelper.MakeResourceActionRoles()

	mockClient.ExpectScan(nil, expRar)
	policies, err := svc.GetPolicies(*app)
	assert.NoError(t, err)
	assert.NotNil(t, policies)
	assert.Equal(t, []rar.ResourceActionRoles{expRar}, policies)
}

func newPolicyStoreSvcAndClient() (policystore.PolicyBackendSvc[testhelper.SimpleDynamodbItem], *testhelper.MockClient) {
	mockClient := testhelper.NewMockClient()
	aOpt := dynamodbpolicystore.WithDynamodbClientOverride[testhelper.SimpleDynamodbItem](mockClient)
	svc, _ := dynamodbpolicystore.NewPolicyStoreSvc(testhelper.SimpleTableInfo(), testhelper.AwsCredentialsForTest(), aOpt)
	return svc, mockClient
}

func newSvcWithDynamicTableDef() (policystore.PolicyBackendSvc[rar.DynamicResourceActionRolesMapper], *testhelper.MockClient) {
	mockClient := testhelper.NewMockClient()
	aOpt := dynamodbpolicystore.WithDynamodbClientOverride[rar.DynamicResourceActionRolesMapper](mockClient)
	svc, _ := dynamodbpolicystore.NewPolicyStoreSvc(testhelper.DynamicTableInfo(), testhelper.AwsCredentialsForTest(), aOpt)
	return svc, mockClient
}
