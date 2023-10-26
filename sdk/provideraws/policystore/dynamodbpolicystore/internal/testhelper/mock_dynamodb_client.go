package testhelper

import (
	"context"
	"encoding/json"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	"github.com/stretchr/testify/mock"
	log "golang.org/x/exp/slog"
)

type MockClient struct {
	mock.Mock
	// Resource, Action, Members attribute names in our dynamodb table
	// e.g. some of our tests use ResourceX, ActionX, MembersX as the column names
	tableDefn dynamodbpolicystore.TableDefinition
}

func NewMockClient(tableDefn dynamodbpolicystore.TableDefinition) *MockClient {
	return &MockClient{tableDefn: tableDefn}
}

func (m *MockClient) Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ddb.ScanOutput), args.Error(1)
}
func (m *MockClient) UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ddb.UpdateItemOutput), args.Error(1)
}

func (m *MockClient) ExpectScan(andRetError error, orRetItems ...rar.ResourceActionRoles) {
	input := &ddb.ScanInput{TableName: &TableName}
	var output *ddb.ScanOutput
	if andRetError == nil {
		output = CustomScanOutputWithAttributeNames(m.tableDefn, orRetItems...)
		//output = CustomScanOutput(orRetItems...)
	}

	m.On("Scan", context.TODO(), input, mock.AnythingOfType("[]func(*dynamodb.Options)")).
		Return(output, andRetError)
}

func (m *MockClient) ExpectUpdateItem(withInput rar.ResourceActionRoles, andRetError error) {
	output := &ddb.UpdateItemOutput{}

	theFunc := mock.MatchedBy(func(input *ddb.UpdateItemInput) bool {
		resource := input.Key[m.tableDefn.ResourceAttrName].(*types.AttributeValueMemberS)
		action := input.Key[m.tableDefn.ActionAttrName].(*types.AttributeValueMemberS)
		members := input.ExpressionAttributeValues[":members"].(*types.AttributeValueMemberS)
		expMembers, _ := json.Marshal(withInput.Members())
		updateExpr := *input.UpdateExpression
		memberExprAttrName := input.ExpressionAttributeNames["#members"]

		ok := *input.TableName == TableName &&
			resource.Value == withInput.Resource() &&
			action.Value == withInput.Actions()[0] && // TODO handle array
			members.Value == string(expMembers) &&
			updateExpr == "SET #members = :members" &&
			memberExprAttrName == m.tableDefn.MembersAttrName

		if !ok {
			log.Error("test", "unexpected UpdateItem tableName", TableName,
				"Resource", resource.Value, "Action", action.Value, "Members", members.Value,
				"UpdateExpression", updateExpr, "ExpressionAttributeNames", memberExprAttrName)
		}

		return ok
	})

	m.On("UpdateItem", context.TODO(), theFunc, mock.AnythingOfType("[]func(*dynamodb.Options)")).
		Return(output, andRetError)
}
