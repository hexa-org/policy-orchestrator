package dynamodbpolicystore

import (
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/table"
	log "golang.org/x/exp/slog"
)

func NewAttributeDefinition(nameOrPath string, valType string, pk bool, sk bool) *table.AttributeDefinition {
	return table.NewAttributeDefinition(nameOrPath, valType, pk, sk)
}

func NewSimpleTableInfo[R rar.ResourceActionRolesMapper](tableName string, sampleItem R) (*table.TableInfo[R], error) {
	return table.NewSimpleTableInfo(tableName, sampleItem)
}

func NewDynamicTableInfo(tableName string, resourceAttrDef, actionsAttrDef, membersAttrDef *table.AttributeDefinition) (*table.TableInfo[rar.DynamicResourceActionRolesMapper], error) {
	return table.NewDynamicTableInfo(tableName, resourceAttrDef, actionsAttrDef, membersAttrDef)
}

func NewTableDefinitionV2(jsonStr string) (table.TableDefinition, error) {
	var defV2 table.TableDefinitionV2
	err := json.Unmarshal([]byte(jsonStr), &defV2)
	if err != nil {
		log.Error("NewTableDefinitionV2", "msg", "failed to marshall string to TableDefinitionV2",
			"jsonStr", jsonStr, "error", err)
		return table.TableDefinitionV2{}, err
	}
	return defV2, nil
}
