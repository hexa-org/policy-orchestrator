package testhelper

import (
	"encoding/json"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
)

const (
	ActionGet    = "GET"
	ResourceHrUs = "/humanresources/us"
	MembersHrUs  = "Read.HRUS"
)

func MakeResourceActionRoles() rar.ResourceActionRoles {
	return CustomResourceActionRoles(ResourceHrUs, ActionGet, []string{MembersHrUs})
}

func CustomResourceActionRoles(res, action string, members []string) rar.ResourceActionRoles {
	aRar, _ := rar.NewResourceActionRoles(res, []string{action}, members)
	return aRar
}

func ScanOutput() *ddb.ScanOutput {
	rar := MakeResourceActionRoles()
	return CustomScanOutput(rar)
}

func CustomScanOutput(rarList ...rar.ResourceActionRoles) *ddb.ScanOutput {
	items := make([]map[string]types.AttributeValue, 0)
	for _, rar := range rarList {
		members, _ := json.Marshal(rar.Members())

		anItem := map[string]types.AttributeValue{
			"ResourceX": &types.AttributeValueMemberS{Value: rar.Resource()},
			"ActionX":   &types.AttributeValueMemberS{Value: rar.Actions()[0]}, // TODO - haldle array
			"MembersX":  &types.AttributeValueMemberS{Value: string(members)},
		}

		items = append(items, anItem)
	}
	output := &ddb.ScanOutput{Items: items}
	return output
}

func CustomScanOutputWithAttributeNames(tableDefinition dynamodbpolicystore.TableDefinition, rarList ...rar.ResourceActionRoles) *ddb.ScanOutput {
	items := make([]map[string]types.AttributeValue, 0)
	for _, rar := range rarList {
		members, _ := json.Marshal(rar.Members())

		anItem := map[string]types.AttributeValue{
			tableDefinition.ResourceAttrName: &types.AttributeValueMemberS{Value: rar.Resource()},
			tableDefinition.ActionAttrName:   &types.AttributeValueMemberS{Value: rar.Actions()[0]}, // TODO - haldle array
			tableDefinition.MembersAttrName:  &types.AttributeValueMemberS{Value: string(members)},
		}

		items = append(items, anItem)
	}
	output := &ddb.ScanOutput{Items: items}
	return output
}
