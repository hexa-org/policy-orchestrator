package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport/awstestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithSimpleTableInfo(t *testing.T) {
	creds := awstestsupport.AwsCredentialsForTest()
	p, err := orchestrator.NewOrchestrationProvider(creds, creds)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestDynamicTableInfo(t *testing.T) {
	creds := awstestsupport.AwsCredentialsForTest()
	resOpt := orchestrator.WithResourceAttrDefinition("ResourceX", "string", true, false)
	actionsOpt := orchestrator.WithActionsAttrDefinition("ActionsX", "string", false, true)
	membersOpt := orchestrator.WithMembersAttrDefinition("MembersX", "string")
	p, err := orchestrator.NewOrchestrationProviderWithDynamicTableInfo(creds, creds, resOpt, actionsOpt, membersOpt)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}
