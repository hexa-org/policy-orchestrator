package dynamodbpolicystore

import (
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	log "golang.org/x/exp/slog"
)

// TableDefinition - // Resource, Action, Members attribute names in the dynamodb table
// Clients provide this based on their dynamodb table definition
type TableDefinition struct {
	ResourceAttrName string
	ActionAttrName   string
	MembersAttrName  string
}

type TableInfo[R rar.ResourceActionRolesMapper] struct {
	TableName          string
	TableDefinition    TableDefinition
	ItemType           R
	ItemMappingDynamic bool // MapTo static, or dynamic. Decides whether to use item.MapTo() or MapTo(interface{})
	TableDefinitionV2  TableDefinitionV2
}

type AttributeDefinition struct {
	NameOrPath string `json:"nameOrPath"`
	ValType    string `json:"valType"`
}

type MetadataKeyInfo struct {
	Attribute string `json:"attribute"`
}

// TableDefinitionV2 TODO change to TableDefinition
// Also refactor into reusable structs
type TableDefinitionV2 struct {
	Metadata struct {
		Pk MetadataKeyInfo `json:"pk"`
		Sk MetadataKeyInfo `json:"sk"`
	} `json:"metadata"`
	Attributes struct {
		Resource AttributeDefinition `json:"resource"`
		Actions  AttributeDefinition `json:"actions"`
		Members  AttributeDefinition `json:"members"`
	} `json:"attributes"`
}

// NewTableDefinitionV2 - TODO validate
// One must be pk, but only one
// If sk defined, can have multiple.
// attributes can be of valType string, int, []string, []int
// "attributes" must have 3 keys i.e. "resource", "actions", "members"

func NewTableDefinitionV2(jsonStr string) (TableDefinitionV2, error) {
	var defV2 TableDefinitionV2
	err := json.Unmarshal([]byte(jsonStr), &defV2)
	if err != nil {
		log.Error("NewTableDefinitionV2", "msg", "failed to marshall string to TableDefinitionV2",
			"jsonStr", jsonStr, "error", err)
		return TableDefinitionV2{}, err
	}
	return defV2, nil
}
