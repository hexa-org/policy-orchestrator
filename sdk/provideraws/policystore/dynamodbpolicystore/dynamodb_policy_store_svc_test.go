package dynamodbpolicystore_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPolicyStoreSvc_Error(t *testing.T) {
	tableInfo := dynamodbpolicystore.TableInfo[testhelper.TestTableItem]{}
	svc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, []byte("$$"))
	assert.ErrorContains(t, err, "invalid character '$'")
	assert.Nil(t, svc)

}

func TestNewPolicyStoreSvc(t *testing.T) {
	tableInfo := dynamodbpolicystore.TableInfo[testhelper.TestTableItem]{
		TableName:       testhelper.TableName,
		TableDefinition: testhelper.TableDefinition(),
		ItemType:        testhelper.TestTableItem{},
	}
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

func newPolicyStoreSvcAndClient() (policystore.PolicyBackendSvc[testhelper.TestTableItem], *testhelper.MockClient) {
	tableItemInfo := testhelper.TestTableItem{}
	tableInfo := dynamodbpolicystore.TableInfo[testhelper.TestTableItem]{
		TableName:       testhelper.TableName,
		TableDefinition: testhelper.TableDefinition(),
		ItemType:        tableItemInfo,
	}

	mockClient := testhelper.NewMockClient(testhelper.TableDefinition())
	opt := dynamodbpolicystore.WithDynamodbClientOverride[testhelper.TestTableItem](mockClient)
	svc, _ := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, testhelper.AwsCredentialsForTest(), opt)
	return svc, mockClient
}
