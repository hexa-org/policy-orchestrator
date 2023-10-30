package dynamodbpolicystore

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/client"
	log "golang.org/x/exp/slog"
)

type PolicyStoreSvc[R rar.ResourceActionRolesMapper] struct {
	client    client.DynamodbClient
	tableInfo TableInfo[R]
}

type Opt[R rar.ResourceActionRolesMapper] func(svc *PolicyStoreSvc[R])

func WithDynamodbClientOverride[R rar.ResourceActionRolesMapper](client client.DynamodbClient) Opt[R] {
	return func(svc *PolicyStoreSvc[R]) {
		svc.client = client
	}
}

func NewPolicyStoreSvc[R rar.ResourceActionRolesMapper](tableInfo TableInfo[R], key []byte, opts ...Opt[R]) (policystore.PolicyBackendSvc[R], error) {
	svc := &PolicyStoreSvc[R]{tableInfo: tableInfo}
	if len(opts) == 0 {
		c, err := client.NewDynamodbClient(key, nil)
		if err != nil {
			return nil, err
		}
		svc.client = c
	}

	for _, o := range opts {
		o(svc)
	}
	return svc, nil
}

func (s *PolicyStoreSvc[R]) GetPolicies(_ idp.AppInfo) ([]rar.ResourceActionRoles, error) {
	input := &ddb.ScanInput{TableName: &s.tableInfo.TableName}
	output, err := s.client.Scan(context.TODO(), input)

	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to Scan table. Err=", err)
		return nil, err
	}

	// If dynamic item json, then use MapTo(interface{} )
	if s.tableInfo.ItemMappingDynamic {
		return getPoliciesDynamic(output, s.toRarList)
	}

	// If can be mapped to a provided struct, then use simple MapTo
	return s.getPoliciesSimple(output, rar.ToResourceActionRoleList[R])
	/*var items []R
	err = attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to unmarshal items. Err=", err)
		return nil, err
	}*/

	//return rar.ToResourceActionRoleList(items)
}

func getPoliciesDynamic(output *ddb.ScanOutput, mapperFunc func(theItems []interface{}) ([]rar.ResourceActionRoles, error)) ([]rar.ResourceActionRoles, error) {
	var items []interface{}
	err := attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to unmarshal items. Err=", err)
		return nil, err
	}
	return mapperFunc(items)
}

func (s *PolicyStoreSvc[R]) getPoliciesSimple(output *ddb.ScanOutput, mapperFunc func(theItems []R) ([]rar.ResourceActionRoles, error)) ([]rar.ResourceActionRoles, error) {
	var items []R
	err := attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to unmarshal items. Err=", err)
		return nil, err
	}
	return mapperFunc(items)
}

func (s *PolicyStoreSvc[R]) SetPolicy(aRar rar.ResourceActionRoles) error {
	log.Info("PolicyStoreSvc.SetPolicy", "msg", "aRar", aRar)

	/*tableDefinition := s.tableInfo.TableDefinition
	inputBuilder := NewInputBuilder(s.tableInfo.TableName, map[string]string{
		"Resource": tableDefinition.ResourceAttrName,
		"Action":   tableDefinition.ActionAttrName,
		"Members":  tableDefinition.MembersAttrName,
	})
	//input, err := inputBuilder.UpdateItemInput(rar)
	*/

	// Update nested items
	// https://repost.aws/questions/QUQxPvh3XLQQeDNUM1s3Y9vA/dynamodb-update-deep-nested-attributes

	/*builderV2 := NewInputBuilderV2(s.tableInfo.TableName, TableDefinitionV2{
		Metadata: struct {
			Pk MetadataKeyInfo `json:"pk"`
			Sk MetadataKeyInfo `json:"sk"`
		}{
			Pk: MetadataKeyInfo{Attribute: "resource"},
			Sk: MetadataKeyInfo{Attribute: "actions"},
		},
		Attributes: struct {
			Resource AttributeDefinition `json:"resource"`
			Actions  AttributeDefinition `json:"actions"`
			Members  AttributeDefinition `json:"members"`
		}{
			Resource: AttributeDefinition{
				NameOrPath: "ResourceX",
				ValType:    "string",
			},
			Actions: AttributeDefinition{
				NameOrPath: "ActionsX",
				ValType:    "string",
			},
			Members: AttributeDefinition{
				NameOrPath: "MembersX",
				ValType:    "string",
			},
		},
	})*/

	builderV2 := NewInputBuilderV2(s.tableInfo.TableName, s.tableInfo.TableDefinitionV2)

	input, err := builderV2.updateItemInputV2(aRar)
	if err != nil {
		log.Error("PolicyStoreSvc.SetPolicy", "msg", "failed to build updateItemInput", "error", err)
		return err
	}

	// TODO - process output
	log.Error("PolicyStoreSvc.SetPolicy", "msg", input)
	_, err = s.client.UpdateItem(context.TODO(), input)
	return err
}

func (s *PolicyStoreSvc[R]) toRarList(scanOutputItems []interface{}) ([]rar.ResourceActionRoles, error) {
	rars := make([]rar.ResourceActionRoles, 0)
	for _, anItem := range scanOutputItems {
		aRar, err := s.tableInfo.ItemType.MapToV2(anItem)
		if err != nil {
			return nil, err
		}
		rars = append(rars, aRar)
	}
	return rars, nil

}
