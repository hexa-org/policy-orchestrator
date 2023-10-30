package dynamodbpolicystore

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

// UpdateItemInput
// nested items update
// https://repost.aws/questions/QUQxPvh3XLQQeDNUM1s3Y9vA/dynamodb-update-deep-nested-attributes
// Primary key attribute must be scalar etc
// The only data types allowed for primary key attributes are string, number, or binary
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.CoreComponents.html#HowItWorks.CoreComponents.PrimaryKey
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

type InputBuilderV2 struct {
	tableName                 string
	tableDefinitionV2         TableDefinitionV2
	keyAttrVal                map[string]types.AttributeValue
	expressionAttributeNames  map[string]string
	expressionAttributeValues map[string]types.AttributeValue
	updateExpression          string
}

func NewInputBuilderV2(tableName string, tableDefinitionV2 TableDefinitionV2) *InputBuilderV2 {
	return &InputBuilderV2{
		tableName:                 tableName,
		tableDefinitionV2:         tableDefinitionV2,
		keyAttrVal:                map[string]types.AttributeValue{},
		expressionAttributeNames:  map[string]string{},
		expressionAttributeValues: map[string]types.AttributeValue{}}
}

func (t2 *InputBuilderV2) UpdateItemInput(rar rar.ResourceActionRoles) (*dynamodb.UpdateItemInput, error) {
	return t2.updateItemInputV2(rar)
}

// updateItemInputV2 - builds input without making any assumption on the attribute names or types
// TODO - 1) add support for composite fields 2) array values
func (t2 *InputBuilderV2) updateItemInputV2(aRar rar.ResourceActionRoles) (*dynamodb.UpdateItemInput, error) {
	log.Info("InputBuilderV2.updateItemInputV2", "aRar", aRar)
	err := t2.makeInput(aRar)
	if err != nil {
		log.Error("updateItemInputV2", "msg", "failed to build updateInputV2", "error", err)
		return nil, err
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                &t2.tableName,
		Key:                      t2.keyAttrVal,
		ExpressionAttributeNames: t2.expressionAttributeNames,
		//map[string]string{	"#members": t.membersAttrName(),},
		ExpressionAttributeValues: t2.expressionAttributeValues,
		//map[string]types.AttributeValue{":members": membersVal,	},
		UpdateExpression: &t2.updateExpression,
		ReturnValues:     types.ReturnValueAllNew,
	}

	return input, nil
}

func (t2 *InputBuilderV2) makeInput(aRar rar.ResourceActionRoles) error {
	log.Info("InputBuilderV2.makeInput BEGIN")

	tableAttributes := t2.tableDefinitionV2.Attributes
	log.Info("InputBuilderV2.updateItemInputV2", "tableAttributes", tableAttributes)

	attrMapInfo := make(map[string]attrDefinitionAndValue)
	attrMapInfo["resource"] = attrDefinitionAndValue{val: []string{aRar.Resource()}, attrDefinition: tableAttributes.Resource}
	attrMapInfo["actions"] = attrDefinitionAndValue{val: aRar.Actions(), attrDefinition: tableAttributes.Actions}

	membersStr, err := json.Marshal(aRar.Members())
	log.Info("InputBuilderV2.updateItemInputV2", "membersStr", membersStr)
	if err != nil {
		log.Error("updateItemInput error marshall member array from", "members", aRar.Members(), "Err", err)
		return err
	}

	attrMapInfo["members"] = attrDefinitionAndValue{val: []string{string(membersStr)}, attrDefinition: tableAttributes.Members}

	log.Info("InputBuilderV2.updateItemInputV2", "attrMapInfo", attrMapInfo)

	// Add primary key, which is required
	keyColName, pkVal, err := makeItemKey(t2.tableDefinitionV2.Metadata.Pk.Attribute, attrMapInfo)
	log.Info("InputBuilderV2.updateItemInputV2", "keyColName", keyColName, "pkVal", pkVal, "error", err)
	if err != nil {
		return err
	}
	t2.addToKey(keyColName, pkVal)

	// Add range key if defined.
	if t2.tableDefinitionV2.Metadata.Sk != *new(MetadataKeyInfo) && t2.tableDefinitionV2.Metadata.Sk.Attribute != "" {
		keyColName, pkVal, err = makeItemKey(t2.tableDefinitionV2.Metadata.Sk.Attribute, attrMapInfo)
		log.Info("InputBuilderV2.updateItemInputV2", "sort key ColName", keyColName, "sortKeyVal", pkVal, "error", err)
		if err != nil {
			return err
		}
		t2.addToKey(keyColName, pkVal)
	}

	colName, val, err := makeItemKey("members", attrMapInfo)
	log.Info("InputBuilderV2.updateItemInputV2", "members ColName", colName, "Val", val, "error", err)
	if err != nil {
		return err
	}
	t2.addToExprNameAndValue(colName, val)
	t2.updateExpression = fmt.Sprintf("SET #%s = :%s", colName, colName)
	return nil
}

func (t2 *InputBuilderV2) addToKey(tableAttrName string, attrValue types.AttributeValue) {
	valPlaceholder, _ := attributevalue.Marshal(":" + tableAttrName)
	t2.keyAttrVal["#"+tableAttrName] = valPlaceholder //pkVal
	t2.addToExprNameAndValue(tableAttrName, attrValue)

}

func (t2 *InputBuilderV2) addToExprNameAndValue(tableAttrName string, attrValue types.AttributeValue) {
	t2.expressionAttributeNames["#"+tableAttrName] = tableAttrName
	t2.expressionAttributeValues[":"+tableAttrName] = attrValue
}

type attrDefinitionAndValue struct {
	val            []string
	attrDefinition AttributeDefinition
}

// makeItemKey
// keyInfo - metadata.pk or medata.sk gives the "key" (resource, actions, members) pointer to the attribute definition
// returns the attribute name, value and/or error
func makeItemKey(attrKey string, attrDefinitionAndVal map[string]attrDefinitionAndValue) (string, types.AttributeValue, error) {
	keyAttrDefinition := attrDefinitionAndVal[attrKey].attrDefinition
	keyVal := attrDefinitionAndVal[attrKey].val

	// TODO Support nested / composite attributes.
	// Attribute definition already accepts any string that can be a path
	nameOrPath := keyAttrDefinition.NameOrPath
	valType := keyAttrDefinition.ValType

	// TODO Support array of int, string
	if valType == "int" || valType == "string" {
		aVal, err := attributevalue.Marshal(keyVal[0])
		if err != nil {
			log.Error("makeItemKey", "msg", "error marshall attribute value",
				"found", valType, "keyAttrKey", attrKey, "nameOrPath", nameOrPath,
				"error", err)
			return "", nil, err
		}

		return nameOrPath, aVal, nil
	} else {
		log.Error("makeItemKey", "msg", "invalid table definition provided. attributes must be either string or int",
			"found", valType, "keyAttrKey", attrKey, "nameOrPath", nameOrPath)
		return "", nil, fmt.Errorf("invalid table definition provided. attributes must be either string or int")
	}
}
