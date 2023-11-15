package testhelper

import (
	"context"
	"encoding/json"
	"fmt"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/stretchr/testify/mock"
	"reflect"
)

type MockClient struct {
	mock.Mock
}

func NewMockClient() *MockClient {
	return &MockClient{}
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
		output = CustomScanOutputWithAttributeNames(orRetItems...)
	}

	m.On("Scan", context.TODO(), input, mock.AnythingOfType("[]func(*dynamodb.Options)")).
		Return(output, andRetError)
}

func (m *MockClient) ExpectUpdateItem(withInput rar.ResourceActionRoles, andRetError error) {
	output := &ddb.UpdateItemOutput{}

	theFunc := mock.MatchedBy(func(input *ddb.UpdateItemInput) bool {

		expMembers, _ := json.Marshal(withInput.Members())
		updateExpr := fmt.Sprintf("SET #%s = :%s", AttrNameMembers, AttrNameMembers)
		keys := map[string]types.AttributeValue{
			AttrNameResource: &types.AttributeValueMemberS{Value: withInput.Resource()},
			AttrNameActions:  &types.AttributeValueMemberS{Value: withInput.Actions()[0]},
		}
		exprNames := map[string]string{
			//AttrNameExprResource: AttrNameResource,
			//AttrNameExprActions:  AttrNameActions,
			AttrNameExprMembers: AttrNameMembers,
		}

		exprValues := map[string]types.AttributeValue{
			//AttrResourcePlaceholder: &types.AttributeValueMemberS{Value: withInput.Resource()},
			//AttrActionsPlaceholder:  &types.AttributeValueMemberS{Value: withInput.Actions()[0]},
			AttrMembersPlaceholder: &types.AttributeValueMemberS{Value: string(expMembers)},
		}

		expUpdateItemInput := &ddb.UpdateItemInput{
			TableName:                 &TableName,
			Key:                       keys,
			ExpressionAttributeNames:  exprNames,
			ExpressionAttributeValues: exprValues,
			UpdateExpression:          &updateExpr,
			ReturnValues:              types.ReturnValueAllNew,
		}
		ok := reflect.DeepEqual(input, expUpdateItemInput)
		return ok
	})

	m.On("UpdateItem", context.TODO(), theFunc, mock.AnythingOfType("[]func(*dynamodb.Options)")).
		Return(output, andRetError)
}
