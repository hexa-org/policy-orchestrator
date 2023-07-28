package dynamodbpolicy_test

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw/dynamodbpolicy"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateItemInput_NoResource(t *testing.T) {
	rar := providerscommon.NewResourceActionRoles("", "GET", nil)
	input, err := dynamodbpolicy.UpdateItemInput(rar)
	assert.ErrorContains(t, err, "empty resource")
	assert.Nil(t, input)

	rar = providerscommon.NewResourceActionRoles("  ", "GET", nil)
	input, err = dynamodbpolicy.UpdateItemInput(rar)
	assert.ErrorContains(t, err, "empty resource")
	assert.Nil(t, input)
}
func TestUpdateItemInput_NoMembers(t *testing.T) {
	rar := providerscommon.NewResourceActionRoles(" /profile ", "GET", nil)
	input, err := dynamodbpolicy.UpdateItemInput(rar)
	assert.NoError(t, err)

	var actVal string
	err = attributevalue.Unmarshal(input.Key["Resource"], &actVal)
	assert.NoError(t, err)
	assert.Equal(t, "/profile", actVal)

	err = attributevalue.Unmarshal(input.Key["Action"], &actVal)
	assert.NoError(t, err)
	assert.Equal(t, "GET", actVal)

	err = attributevalue.Unmarshal(input.ExpressionAttributeValues[":members"], &actVal)
	assert.NoError(t, err)
	assert.Equal(t, "[]", actVal)
}

func TestUpdateItemInput_WithMembers(t *testing.T) {
	rar := providerscommon.NewResourceActionRoles("   /profile   ", "GET", []string{"  ", " mem1", " mem2 "})
	input, err := dynamodbpolicy.UpdateItemInput(rar)
	assert.NoError(t, err)

	var actVal string
	err = attributevalue.Unmarshal(input.Key["Resource"], &actVal)
	assert.NoError(t, err)
	assert.Equal(t, "/profile", actVal)

	err = attributevalue.Unmarshal(input.Key["Action"], &actVal)
	assert.NoError(t, err)
	assert.Equal(t, "GET", actVal)

	err = attributevalue.Unmarshal(input.ExpressionAttributeValues[":members"], &actVal)
	assert.NoError(t, err)
	assert.Equal(t, `["mem1","mem2"]`, actVal)
}
