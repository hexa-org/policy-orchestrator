package dynamodbpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw/dynamodbpolicy"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/slices"
	"testing"
)

func TestGetResourceRoles_ErrorScan(t *testing.T) {
	client := newMockDynamodbClient()
	client.expectScan(nil, errors.New("some error"))
	svc := dynamodbpolicy.NewPolicyStoreSvc(client)
	rarList, err := svc.GetResourceRoles()
	assert.ErrorContains(t, err, "some error")
	assert.Nil(t, rarList)
}

func TestGetResourceRoles_EmptyRespError(t *testing.T) {
	client := newMockDynamodbClient()
	client.expectScan(nil, nil)
	svc := dynamodbpolicy.NewPolicyStoreSvc(client)
	rarList, err := svc.GetResourceRoles()
	assert.ErrorContains(t, err, "empty")
	assert.Empty(t, rarList)
}

func TestGetResourceRoles_UnmarshallError(t *testing.T) {
	client := newMockDynamodbClient()
	anItem := map[string]types.AttributeValue{
		"ResourceX": &types.AttributeValueMemberS{Value: "something"},
		"ActionX":   &types.AttributeValueMemberS{Value: "GET"},
		"MembersX":  &types.AttributeValueMemberS{Value: `["some-role"]`},
	}

	items := []map[string]types.AttributeValue{anItem}
	output := &ddb.ScanOutput{Items: items}
	client.expectScan(output, nil)
	svc := dynamodbpolicy.NewPolicyStoreSvc(client)
	rarList, err := svc.GetResourceRoles()
	assert.ErrorContains(t, err, "unexpected end of JSON input")
	assert.Nil(t, rarList)
}

func TestGetResourceRoles_UnmarshallMembersError(t *testing.T) {
	client := newMockDynamodbClient()
	anItem := map[string]types.AttributeValue{
		"Resource": &types.AttributeValueMemberS{Value: "something"},
		"Action":   &types.AttributeValueMemberS{Value: "GET"},
		"Members":  &types.AttributeValueMemberS{Value: "some-role"},
	}

	items := []map[string]types.AttributeValue{anItem}
	output := &ddb.ScanOutput{Items: items}
	client.expectScan(output, nil)
	svc := dynamodbpolicy.NewPolicyStoreSvc(client)
	rarList, err := svc.GetResourceRoles()
	assert.ErrorContains(t, err, "invalid character 's'")
	assert.Nil(t, rarList)
}

func TestGetResourceRoles_WithResourceRoles(t *testing.T) {
	client := newMockDynamodbClient()
	existingRars := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}

	output := scanOutput(existingRars)
	client.expectScan(output, nil)
	svc := dynamodbpolicy.NewPolicyStoreSvc(client)
	rarList, err := svc.GetResourceRoles()
	assert.NoError(t, err)
	assert.NotNil(t, rarList)
	assert.Len(t, rarList, 2)
	for _, actRar := range rarList {
		actionRes := actRar.Action + actRar.Resource
		expMembers := existingRars[actionRes]
		assert.NotEmpty(t, expMembers)
		slices.Sort(expMembers)
		slices.Sort(actRar.Roles)
		assert.Equal(t, expMembers, actRar.Roles)
	}
}

func TestPolicyStoreSvc_UpdateResourceRole_UpdateInputError(t *testing.T) {
	rar := providerscommon.NewResourceActionRoles("  ", "GET", []string{})
	svc := dynamodbpolicy.NewPolicyStoreSvc(nil)
	err := svc.UpdateResourceRole(rar)
	assert.ErrorContains(t, err, "empty resource")
}

func TestPolicyStoreSvc_UpdateResourceRole_DynamodbUpdateItemError(t *testing.T) {
	client := newMockDynamodbClient()
	rar := policytestsupport.MakeRar(policytestsupport.ActionGetHrUs, []string{"some-hr-role"})
	input, _ := dynamodbpolicy.UpdateItemInput(rar)
	client.expectUpdateItem(input, errors.New("some-error"))
	svc := dynamodbpolicy.NewPolicyStoreSvc(client)
	err := svc.UpdateResourceRole(rar)
	assert.ErrorContains(t, err, "some-error")
}

func TestPolicyStoreSvc_UpdateResourceRole_Success(t *testing.T) {
	client := newMockDynamodbClient()
	rar := policytestsupport.MakeRar(policytestsupport.ActionGetHrUs, []string{"some-hr-role"})
	client.expectUpdateItem(dynamodbpolicy.UpdateItemInput(rar))
	svc := dynamodbpolicy.NewPolicyStoreSvc(client)
	err := svc.UpdateResourceRole(rar)
	assert.NoError(t, err)
}

type mockDynamodbClient struct {
	mock.Mock
}

func newMockDynamodbClient() *mockDynamodbClient {
	return &mockDynamodbClient{}
}

func (m *mockDynamodbClient) Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ddb.ScanOutput), args.Error(1)
}

func (m *mockDynamodbClient) UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ddb.UpdateItemOutput), args.Error(1)
}

func (m *mockDynamodbClient) expectScan(output *ddb.ScanOutput, err error) {
	input := &ddb.ScanInput{TableName: &dynamodbpolicy.AwsPolicyStoreTableName}
	m.On("Scan", context.TODO(), input, mock.AnythingOfType("[]func(*dynamodb.Options)")).
		Return(output, err)
}

func (m *mockDynamodbClient) expectUpdateItem(params *ddb.UpdateItemInput, err error) {
	output := &ddb.UpdateItemOutput{}
	m.On("UpdateItem", context.TODO(), params, mock.AnythingOfType("[]func(*dynamodb.Options)")).
		Return(output, err)
}

func scanOutput(rars map[string][]string) *ddb.ScanOutput {
	items := make([]map[string]types.AttributeValue, 0)
	expRars := policytestsupport.MakeRarList(rars)
	for _, rar := range expRars {
		rolesStr, _ := json.Marshal(rar.Roles)
		anItem := map[string]types.AttributeValue{
			"Resource": &types.AttributeValueMemberS{Value: rar.Resource},
			"Action":   &types.AttributeValueMemberS{Value: rar.Action},
			"Members":  &types.AttributeValueMemberS{Value: string(rolesStr)},
		}
		items = append(items, anItem)
	}

	return &ddb.ScanOutput{Items: items}
}
