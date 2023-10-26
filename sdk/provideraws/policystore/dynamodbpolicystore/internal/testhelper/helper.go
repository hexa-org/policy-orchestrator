package testhelper

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/client"
)

const TestAwsRegion = "us-west-1"
const TestAwsAccessKeyId = "anAccessKeyID"
const TestAwsSecretAccessKey = "aSecretAccessKey"

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

func TableDefinition() dynamodbpolicystore.TableDefinition {
	return dynamodbpolicystore.TableDefinition{
		ResourceAttrName: "ResourceX",
		ActionAttrName:   "ActionX",
		MembersAttrName:  "MembersX",
	}
}

func InputBuilder() *client.InputBuilder {
	return client.NewInputBuilder(TableName, map[string]string{
		"Resource": "ResourceX",
		"Action":   "ActionX",
		"Members":  "MembersX",
	})
}
