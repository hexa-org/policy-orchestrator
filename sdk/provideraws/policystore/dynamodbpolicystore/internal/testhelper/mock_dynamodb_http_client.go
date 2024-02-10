package testhelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/stretchr/testify/mock"
	log "golang.org/x/exp/slog"
	"io"
	"net/http"
	"reflect"
)

type MockDynamodbHttpClient struct {
	mock.Mock
}

func NewMockDynamodbHttpClient() *MockDynamodbHttpClient {
	return &MockDynamodbHttpClient{}
}

func (m *MockDynamodbHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	resp := args.Get(0).(*http.Response)
	return resp, args.Error(1)
}

type scanOutputType struct {
	Items []map[string]interface{}
}

// The Key and ExpressionAttributeValues in below struct are actually
// map[string]types.AttributeValue which can hold values of different types
// but during json Unmarshall to map[string]types.AttributeValue doesn't seem to work
// because types.AttributeValue is some sort of generic implementation of aws
// The concrete types are actually types.AttributeValueMemberS (string) etc.
// Instead of interface{}, using types.AttributeValueMemberS works fine
// but then we are restricted to only string values.
// The only way around this I found was using interface{}
// The one caveat is that if you need to access the actual data value, you will need to cast it
// In our case, we are just comparing two structs using Deep equals so it works
type updateItemInputType struct {
	TableName                 string
	Key                       map[string]interface{}
	UpdateExpression          string
	ConditionExpression       string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]interface{}
	ReturnValues              string
}

func (m *MockDynamodbHttpClient) ExpectScan(withRespErr error, orRespItems ...rar.ResourceActionRoles) {
	itemArr := makeDynamodbResourcePolicyItems(orRespItems...)
	itemsResp, err := json.Marshal(itemArr)
	if err != nil {
		log.Error("test", "mock.ExpectScan failed to marshall ScanOutput items", err)
		return
	}
	m.expectCall("DynamoDB_20120810.Scan", withRespErr, itemsResp)
}

func (m *MockDynamodbHttpClient) ExpectUpdateItem(withReq rar.ResourceActionRoles, respWithErr error) {
	theFunc := mock.MatchedBy(func(req *http.Request) bool {
		ok := req.Method == http.MethodPost &&
			req.Header.Get("X-Amz-Target") == "DynamoDB_20120810.UpdateItem"
		if !ok {
			log.Error("test", "mock.ExpectUpdateItem", "expecting dynamodb UpdateItem", "req.Method", req.Method, "X-Amz-Target", req.Header.Get("X-Amz-Target"))
			return false
		}

		// Actual Request
		actBody, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error("test", "mock.ExpectUpdateItem error reading mocked request body", err)
			return false
		}
		var actInput updateItemInputType
		err = json.Unmarshal(actBody, &actInput)
		if err != nil {
			log.Error("test", "mock.ExpectUpdateItem error unmarshall actual request body for UpdateItemInput", err)
			return false
		}

		// Build Expected Request
		membersEscaped := make([]string, 0)
		for _, mem := range withReq.Members() {
			escMem := fmt.Sprintf(`\"%s\"`, mem)
			membersEscaped = append(membersEscaped, escMem)
		}

		membersStr, err := json.Marshal(withReq.Members())

		expInput := updateItemInputType{
			TableName: TableName,
			Key: map[string]interface{}{
				AttrNameResource: ddbTypeString(withReq.Resource()),
				AttrNameActions:  ddbTypeString(withReq.Actions()[0]), // TODO handle array
			},
			UpdateExpression: fmt.Sprintf("SET %s = %s", AttrNameExprMembers, AttrMembersPlaceholder),
			//ConditionExpression:       "",
			ExpressionAttributeNames: map[string]string{AttrNameExprMembers: AttrNameMembers},
			ExpressionAttributeValues: map[string]interface{}{
				AttrMembersPlaceholder: ddbTypeString(string(membersStr)),
			},
			ReturnValues: "ALL_NEW",
		}
		ok = reflect.DeepEqual(expInput, actInput)
		return ok
	})

	statusCode := http.StatusOK
	if respWithErr != nil {
		statusCode = http.StatusBadRequest
	}
	resp := &http.Response{
		StatusCode: statusCode, Body: http.NoBody}
	m.On("Do", theFunc).Return(resp, respWithErr)
}

func (m *MockDynamodbHttpClient) expectCall(amzTarget string, withRespErr error, orRespBytes []byte) {
	var resp *http.Response
	if withRespErr != nil {
		resp = &http.Response{
			StatusCode: http.StatusBadRequest, Body: http.NoBody}
	} else {
		resp = &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(orRespBytes)),
		}
	}

	theFunc := mock.MatchedBy(func(req *http.Request) bool {
		ok := req.Method == http.MethodPost &&
			req.Header.Get("X-Amz-Target") == amzTarget
		return ok
	})

	m.On("Do", theFunc).Return(resp, withRespErr)
}

// makeDynamodbResourcePolicyItem - builds the Items response returned in response to ddb.Scan
// Marshall / Unmarshall does not seem to work, so we have to build the resp with string manipulation.
func makeDynamodbResourcePolicyItems(rarList ...rar.ResourceActionRoles) scanOutputType {

	if len(rarList) == 0 {
		return scanOutputType{}
	}

	items := make([]map[string]interface{}, 0)
	for _, rar := range rarList {
		item := make(map[string]interface{})
		item[AttrNameResource] = ddbTypeString(rar.Resource())
		item[AttrNameActions] = ddbTypeString(rar.Actions()[0]) // TODO handle array
		members := fmt.Sprintf("[\"%s\"]", rar.Members()[0])
		item[AttrNameMembers] = ddbTypeString(members)
		items = append(items, item)
	}

	return scanOutputType{Items: items}
}

func ddbTypeString(val string) map[string]interface{} {
	return map[string]interface{}{
		"S": val,
	}
}
