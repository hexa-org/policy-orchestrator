package testhelper

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
)

const TestAwsRegion = "us-west-1"
const TestAwsAccessKeyId = "anAccessKeyID"
const TestAwsSecretAccessKey = "aSecretAccessKey"

const AttrNameResource = "ResourceX"
const AttrNameActions = "ActionsX"
const AttrNameMembers = "MembersX"

const AttrNameExprResource = "#ResourceX"
const AttrNameExprActions = "#ActionsX"
const AttrNameExprMembers = "#MembersX"

const AttrResourcePlaceholder = ":ResourceX"
const AttrActionsPlaceholder = ":ActionsX"
const AttrMembersPlaceholder = ":MembersX"

var TableName = "TestDynamodbTable"

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

type TestTableItem struct {
	ResourceX string `json:"ResourceX"`
	ActionX   string `json:"ActionX"`
	MembersX  string `json:"MembersX"`
}

func (it TestTableItem) MapTo() (rar.ResourceActionRoles, error) {
	members := make([]string, 0)
	_ = json.Unmarshal([]byte(it.MembersX), &members)
	return rar.NewResourceActionRoles(it.ResourceX, []string{it.ActionX}, members)
}

func (it TestTableItem) MapToV2(item interface{}) (rar.ResourceActionRoles, error) {
	theMap := item.(map[string]interface{})

	aRes := theMap["ResourceX"].(string)
	anAct := theMap["ActionX"].(string)
	aMemStr := theMap["MembersX"].(string)

	members := make([]string, 0)
	_ = json.Unmarshal([]byte(aMemStr), &members)
	return rar.NewResourceActionRoles(aRes, []string{anAct}, members)
}

func TableDefinition() dynamodbpolicystore.TableDefinition {
	return dynamodbpolicystore.TableDefinition{
		ResourceAttrName: "ResourceX",
		ActionAttrName:   "ActionX",
		MembersAttrName:  "MembersX",
	}
}

func InputBuilder() *dynamodbpolicystore.InputBuilder {
	return dynamodbpolicystore.NewInputBuilder(TableName, map[string]string{
		"Resource": "ResourceX",
		"Action":   "ActionX",
		"Members":  "MembersX",
	})
}
