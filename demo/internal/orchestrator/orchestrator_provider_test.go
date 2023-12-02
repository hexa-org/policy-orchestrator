package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2/apps/aws/providercognito"
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2/policy/aws/providerdynamodb"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport/awstestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithSimpleTableInfo(t *testing.T) {
	creds := awstestsupport.AwsCredentialsForTest()
	idpOpt := providercognito.NewCognitoIdp(creds)
	policyStore := providerdynamodb.NewSimpleItemStore("tableName", creds)
	p, err := orchestrator.NewOrchestrationProvider(idpOpt, policyStore)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestDynamicTableInfo(t *testing.T) {
	creds := awstestsupport.AwsCredentialsForTest()
	idpOpt := providercognito.NewCognitoIdp(creds)
	resAttrDef := providerdynamodb.NewAttributeDefinition("ResourceX", "string", true, false)
	actionsAttrDef := providerdynamodb.NewAttributeDefinition("ActionsX", "string", false, true)
	membersDef := providerdynamodb.NewAttributeDefinition("MembersX", "string", false, false)
	tableDef := providerdynamodb.NewTableDefinition(resAttrDef, actionsAttrDef, membersDef)
	policyStore := providerdynamodb.NewDynamicItemStore(providerdynamodb.AwsPolicyStoreTableName, creds, tableDef)
	p, err := orchestrator.NewOrchestrationProvider(idpOpt, policyStore)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}
