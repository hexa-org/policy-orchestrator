package client

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	log "golang.org/x/exp/slog"
	"strings"
)

type InputBuilder struct {
	tableName       string
	tableDefinition map[string]string
}

func NewInputBuilder(tableName string, tableDefinition map[string]string) *InputBuilder {
	return &InputBuilder{tableName: tableName, tableDefinition: tableDefinition}
}

func (t *InputBuilder) resourceAttrName() string {
	return t.tableDefinition["Resource"]
}
func (t *InputBuilder) actionAttrName() string {
	return t.tableDefinition["Action"]
}
func (t *InputBuilder) membersAttrName() string {
	return t.tableDefinition["Members"]
}

func (t *InputBuilder) UpdateItemInput(rar rar.ResourceActionRoles) (*dynamodb.UpdateItemInput, error) {
	return t.updateItemInput(rar)
}

func (t *InputBuilder) updateItemInput(rar rar.ResourceActionRoles) (*dynamodb.UpdateItemInput, error) {
	resource := rar.Resource()
	action := rar.Actions()[0] // TODO handle array
	if strings.TrimSpace(resource) == "" || strings.TrimSpace(action) == "" {
		return nil, fmt.Errorf("empty resource='%s' or action='%s'", resource, action)
	}

	membersVal, err := membersAttributeValue(rar.Members())
	if err != nil {
		return nil, err
	}

	//tableDefn := t.ItemType.TableDefinition()
	aResource, _ := attributevalue.Marshal(strings.TrimSpace(resource))
	anAction, _ := attributevalue.Marshal(strings.TrimSpace(action))
	keyAttrVal := map[string]types.AttributeValue{
		t.resourceAttrName(): aResource,
		t.actionAttrName():   anAction,
	}

	updateExpr := "SET #members = :members"
	input := &dynamodb.UpdateItemInput{
		TableName: &t.tableName,
		Key:       keyAttrVal,
		ExpressionAttributeNames: map[string]string{
			"#members": t.membersAttrName(),
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":members": membersVal,
		},
		UpdateExpression: &updateExpr,
		ReturnValues:     types.ReturnValueAllNew,
	}

	return input, nil
}

func membersAttributeValue(members []string) (types.AttributeValue, error) {
	membersStr, err := json.Marshal(members)
	if err != nil {
		log.Error("updateItemInput error marshall member array from", "members", members, "Err", err)
		return nil, err
	}

	membersVal, err := attributevalue.Marshal(string(membersStr))
	if err != nil {
		log.Error("updateItemInput error building AttributeValue from", "membersStr", membersStr, "Err", err)
		return nil, err
	}
	return membersVal, nil
}
