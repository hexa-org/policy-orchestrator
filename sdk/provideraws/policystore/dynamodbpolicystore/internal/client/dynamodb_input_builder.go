package client

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/table"
	log "golang.org/x/exp/slog"
	"strings"
)

// UpdateItemInput
// nested items update
// https://repost.aws/questions/QUQxPvh3XLQQeDNUM1s3Y9vA/dynamodb-update-deep-nested-attributes
// Primary key attribute must be scalar etc
// The only data types allowed for primary key attributes are string, number, or binary
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.CoreComponents.html#HowItWorks.CoreComponents.PrimaryKey

type InputBuilderV2 struct {
	tableName                 string
	tableDefinition           table.TableDefinition
	keyAttrVal                map[string]types.AttributeValue
	expressionAttributeNames  map[string]string
	expressionAttributeValues map[string]types.AttributeValue
	updateExpression          string
}

func NewInputBuilderV2(tableName string, tableDefinition table.TableDefinition) *InputBuilderV2 {
	// Validate has pk.
	// Only one pk
	// If sk exists, only one sk
	// Validate types "int", "string", "[]int", "[]string"
	return &InputBuilderV2{
		tableName:                 tableName,
		tableDefinition:           tableDefinition,
		keyAttrVal:                map[string]types.AttributeValue{},
		expressionAttributeNames:  map[string]string{},
		expressionAttributeValues: map[string]types.AttributeValue{}}
}

// UpdateItemInput - builds input without making any assumption on the attribute names or types
// TODO - 1) add support for composite fields 2) array values
func (t2 *InputBuilderV2) UpdateItemInput(aRar rar.ResourceActionRoles) (*dynamodb.UpdateItemInput, error) {
	log.Info("InputBuilderV2.updateItemInputV2", "aRar", aRar)
	err := t2.makeInput(aRar)
	if err != nil {
		log.Error("updateItemInputV2", "msg", "failed to build updateInputV2", "error", err)
		return nil, err
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 &t2.tableName,
		Key:                       t2.keyAttrVal,
		ExpressionAttributeNames:  t2.expressionAttributeNames,
		ExpressionAttributeValues: t2.expressionAttributeValues,
		UpdateExpression:          &t2.updateExpression,
		ReturnValues:              types.ReturnValueAllNew,
	}

	return input, nil
}

func (t2 *InputBuilderV2) makeInput(aRar rar.ResourceActionRoles) error {
	log.Info("InputBuilderV2.makeInput BEGIN")
	tDef := t2.tableDefinition
	var aDef *table.AttributeDefinition
	var rarVal []string
	for _, synAttr := range []string{"resource", "actions", "members"} {
		switch synAttr {
		case "resource":
			// seems very odd that resources can be a sk and actions is the pk
			if tDef.ResourceAttrDefinition().Sk || tDef.ResourceAttrDefinition().Pk {
				aDef = tDef.ResourceAttrDefinition()
				rarVal = []string{aRar.Resource()}
			}
			break
		case "actions":
			if tDef.ActionsAttrDefinition().Sk || tDef.ActionsAttrDefinition().Pk {
				aDef = tDef.ActionsAttrDefinition()
				// TODO - only first element of list is accepted.
				// If not defined as key, allow processing as a dynamodb list
				rarVal = aRar.Actions()
			}
			break
		default:
			aDef = tDef.MembersAttrDefinition()
			membersStr, err := json.Marshal(aRar.Members())
			log.Info("InputBuilderV2.updateItemInputV2", "membersStr", membersStr)
			if err != nil {
				log.Error("updateItemInput error marshall member array from", "members", aRar.Members(), "Err", err)
				return err
			}
			// assuming members are stored as a string.
			// TODO - process actual dynamodb list
			rarVal = []string{string(membersStr)}
			//t2.updateExpression = fmt.Sprintf("SET #%s = :%s", aDef.NameOrPath, aDef.NameOrPath)
			useNameOrPath := strings.ReplaceAll(aDef.NameOrPath, "/", ".#")
			t2.updateExpression = fmt.Sprintf("SET #%s = :%s", useNameOrPath, aDef.NameOrPath)
		}

		// aDef can be nil if actions is incorrectly defined, or not defined in the table def
		// we only necessarily need one key attribute from resources, actions
		// though we can accept one as a pk and other as sk
		// but if user specifies only the pk
		// then the other attr must be ignored.
		if aDef != nil {
			log.Info("InputBuilderV2.updateItemInputV2", "processing", synAttr, "attribute", aDef.NameOrPath, "value", rarVal)
			err := t2.addToKeyOrNameValueExpressions(aDef, rarVal)
			if err != nil {
				log.Error("updateItemInput error from makeOne", "attribute", aDef.NameOrPath, "value", rarVal, "Err", err)
				return err
			}
		} else {
			log.Info("InputBuilderV2.updateItemInputV2", "processing", synAttr, "ignore", "not defined as key OR not member")
		}

	}

	return nil
}

func (t2 *InputBuilderV2) addToKeyOrNameValueExpressions(aDef *table.AttributeDefinition, rarVal []string) error {
	nameOrPath := aDef.NameOrPath
	valType := aDef.ValType
	log.Info("InputBuilderV2.makeOne", "nameOrPath", nameOrPath, "value", rarVal, "valType", valType, "pk", aDef.Pk, "sk", aDef.Sk)

	aVal, err := marshallVal(valType, rarVal)
	if err != nil {
		log.Error("makeItemKey", "msg", "error marshall attribute value",
			"nameOrPath", nameOrPath, "valType", valType, "rarVal", rarVal,
			"error", err)
		return err
	}

	if aDef.Pk || aDef.Sk {
		t2.addToKey(nameOrPath, aVal)
	} else {
		t2.addToExprNameAndValue(nameOrPath, aVal)
	}

	return nil
}

func (t2 *InputBuilderV2) addToKey(tableAttrName string, attrValue types.AttributeValue) {
	// key cannot be composite
	t2.keyAttrVal[tableAttrName] = attrValue
}

func (t2 *InputBuilderV2) addToExprNameAndValue(tableAttrName string, attrValue types.AttributeValue) {
	// even though member attr can be composite, this func just uses the leaf
	// attribute name in ExpressionAttributeNames,ExpressionAttributeValues
	// This func will not be used for Action or Resource because, either these are defined as pk, sk
	// or are not used in the expression at all
	// e.g. we need the pk defined, sk is optional
	// so this is invalid config Resource: pk, Actions: "", Member: [mem1, mem2]
	// because Actions is not defined as a key attr and we DON'T update Actions

	// We also don't perform any delete operations.
	// If user deletes a policy from the UI IDQL during SetPolicy
	// we simply set the members = [] (but DO NOT perform delete operations).

	// The whole point of this entire explanation
	// TODO: Validate TableDefinition
	// 1) at least one scalar attr defined as pk
	// 2) member is required
	// 3) only member attr can be composite
	// 4) nameOrPath - only allowed ['/', 'a-zA-Z', '0-9', '_', '-', '.']
	// 5A) first split name or path by '/'
	//		'.'' NOT allowed prefix/suffix

	// If tableDefinition was validated at creation time, we should be good here without any chcecking.
	// This func will only run for members
	t2.expressionAttributeNames["#"+tableAttrName] = tableAttrName
	t2.expressionAttributeValues[":"+tableAttrName] = attrValue
}

func marshallVal(valType string, val []string) (types.AttributeValue, error) {
	// TODO - process []int, []string as well for non-key attributes
	// even composite attributes
	if valType == "int" || valType == "string" {
		aVal, err := attributevalue.Marshal(val[0])
		if err != nil {
			log.Error("marshallVal", "msg", "error marshall attribute value",
				"found", valType, "error", err)
			return nil, err
		}

		return aVal, nil
	} else {
		log.Error("makeItemKey", "msg", "invalid table definition provided. attributes must be either string or int",
			"found", valType)
		return nil, fmt.Errorf("invalid table definition provided. attributes must be either string or int")
	}
}
