package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport/awstestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithSimpleTableInfo(t *testing.T) {
	creds := awstestsupport.AwsCredentialsForTest()
	idpOpt := providersV2.NewCognitoIdp(creds)
	policyStore := providersV2.NewSimpleItemStore("tableName", creds)
	p, err := orchestrator.NewOrchestrationProvider("amazon", idpOpt, policyStore)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestDynamicTableInfo(t *testing.T) {
	creds := awstestsupport.AwsCredentialsForTest()
	idpOpt := providersV2.NewCognitoIdp(creds)
	resAttrDef := providersV2.NewAttributeDefinition("ResourceX", "string", true, false)
	actionsAttrDef := providersV2.NewAttributeDefinition("ActionsX", "string", false, true)
	membersDef := providersV2.NewAttributeDefinition("MembersX", "string", false, false)
	tableDef := providersV2.NewTableDefinition(resAttrDef, actionsAttrDef, membersDef)
	policyStore := providersV2.NewDynamicItemStore(providersV2.AwsPolicyStoreTableName, creds, tableDef)
	p, err := orchestrator.NewOrchestrationProvider("amazon", idpOpt, policyStore)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}
