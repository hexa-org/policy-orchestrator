package testhelper

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/table"
)

const TestAwsRegion = "us-west-1"
const TestAwsAccessKeyId = "anAccessKeyID"
const TestAwsSecretAccessKey = "aSecretAccessKey"

const AttrNameResource = "ResourceX"
const AttrNameActions = "ActionsX"

const AttrNameMembers = "MembersX"
const AttrNameExprMembers = "#MembersX"
const AttrMembersPlaceholder = ":MembersX"

var TableName = "TestDynamodbTable"

const DynamicTableDefinitionJson = `
			{
				"metadata": {
					"pk": { "attribute": "resource" },
					"sk": { "attribute": "actions" }
				},
				"attributes": {
					"resource": { "nameOrPath": "ResourceX", "valType": "string", "pk": true },
					"actions": { "nameOrPath": "ActionsX", "valType": "string", "sk": true },
					"members": { "nameOrPath": "MembersX", "valType": "string" }
				}
			}`

func AwsCredentialsForTest() []byte {
	str := fmt.Sprintf(`
{
  "accessKeyID": "%s",
  "secretAccessKey": "%s",
  "region": "%s"
}
`, TestAwsAccessKeyId, TestAwsSecretAccessKey, TestAwsRegion)

	return []byte(str)
}

type SimpleDynamodbItem struct {
	ResourceX string `json:"ResourceX" meta:"resource,pk"`
	ActionsX  string `json:"ActionsX" meta:"actions,sk"`
	MembersX  string `json:"MembersX" meta:"members"`
}

func (it SimpleDynamodbItem) MapTo() (rar.ResourceActionRoles, error) {
	members := make([]string, 0)
	_ = json.Unmarshal([]byte(it.MembersX), &members)
	return rar.NewResourceActionRoles(it.ResourceX, []string{it.ActionsX}, members)
}

func SimpleTableInfo() *table.TableInfo[SimpleDynamodbItem] {
	tableInfo, _ := table.NewSimpleTableInfo(TableName, SimpleDynamodbItem{})
	return tableInfo
}

func AttributeDefinitions() (resDef, actionsDef, membersDef *table.AttributeDefinition) {
	resDef = table.NewAttributeDefinition(AttrNameResource, "string", true, false)
	actionsDef = table.NewAttributeDefinition(AttrNameActions, "string", false, true)
	membersDef = table.NewAttributeDefinition(AttrNameMembers, "string", false, false)
	return

}

func DynamicTableInfo() *table.TableInfo[rar.DynamicResourceActionRolesMapper] {
	resDef, actionsDef, membersDef := AttributeDefinitions()
	tableInfo, _ := table.NewDynamicTableInfo(TableName, resDef, actionsDef, membersDef)
	return tableInfo
}

func DynamicTableDefinition() table.TableDefinition {
	resourceAttrDef, actionsAttrDef, membersAttrDef := AttributeDefinitions()
	return table.TableDefinitionV2{
		Attributes: table.TableAttributes{
			Resource: resourceAttrDef,
			Actions:  actionsAttrDef,
			Members:  membersAttrDef,
		},
	}
}
