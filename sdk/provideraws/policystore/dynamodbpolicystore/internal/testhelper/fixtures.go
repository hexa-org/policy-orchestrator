package testhelper

import (
	"encoding/json"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
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
	aRar := MakeResourceActionRoles()
	return customScanOutput(aRar)
}

func customScanOutput(rarList ...rar.ResourceActionRoles) *ddb.ScanOutput {
	items := make([]map[string]types.AttributeValue, 0)
	for _, aRar := range rarList {
		members, _ := json.Marshal(aRar.Members())

		anItem := map[string]types.AttributeValue{
			AttrNameResource: &types.AttributeValueMemberS{Value: aRar.Resource()},
			AttrNameActions:  &types.AttributeValueMemberS{Value: aRar.Actions()[0]}, // TODO - haldle array
			AttrNameMembers:  &types.AttributeValueMemberS{Value: string(members)},
		}

		items = append(items, anItem)
	}
	output := &ddb.ScanOutput{Items: items}
	return output
}

func CustomScanOutputWithAttributeNames(rarList ...rar.ResourceActionRoles) *ddb.ScanOutput {
	items := make([]map[string]types.AttributeValue, 0)
	for _, aRar := range rarList {
		members, _ := json.Marshal(aRar.Members())

		anItem := map[string]types.AttributeValue{
			AttrNameResource: &types.AttributeValueMemberS{Value: aRar.Resource()},
			AttrNameActions:  &types.AttributeValueMemberS{Value: aRar.Actions()[0]}, // TODO - haldle array
			AttrNameMembers:  &types.AttributeValueMemberS{Value: string(members)},
		}

		items = append(items, anItem)
	}
	output := &ddb.ScanOutput{Items: items}
	return output
}
