package table

const PolicyAttrTypeResource = "resource"
const PolicyAttrTypeActions = "actions"
const PolicyAttrTypeMembers = "members"

//var PolicyAttrTypes = []string{PolicyAttrTypeResource, PolicyAttrTypeActions, PolicyAttrTypeMembers}

type TableDefinition interface {
	ResourceAttrDefinition() *AttributeDefinition
	ActionsAttrDefinition() *AttributeDefinition
	MembersAttrDefinition() *AttributeDefinition
	AttrDefinitionMap() map[string]*AttributeDefinition
}

type MetadataKeyInfo struct {
	Attribute string `json:"attribute"`
}

// TableDefinitionV2 - dynamic table definition provided by consumers
// Also refactor into reusable structs
type TableMetadata struct {
	Pk MetadataKeyInfo `json:"pk"`
	Sk MetadataKeyInfo `json:"sk"`
}
type TableAttributes struct {
	Resource *AttributeDefinition `json:"resource"`
	Actions  *AttributeDefinition `json:"actions"`
	Members  *AttributeDefinition `json:"members"`
}

type TableDefinitionV2 struct {
	Metadata   TableMetadata   `json:"metadata"`
	Attributes TableAttributes `json:"attributes"`
}

/*
func NewTableDefinition(attrMap map[string]*table.AttributeDefinition) TableDefinition {
	tableDef := &TableDefinitionV2{
		Attributes: TableAttributes{
			Resource: attrMap[PolicyAttrTypeResource],
			Actions:  attrMap[PolicyAttrTypeActions],
			Members:  attrMap[PolicyAttrTypeMembers],
		},
	}
	return tableDef
}
*/

func (t TableDefinitionV2) ResourceAttrDefinition() *AttributeDefinition {
	return t.Attributes.Resource
}

func (t TableDefinitionV2) ActionsAttrDefinition() *AttributeDefinition {
	return t.Attributes.Actions
}

func (t TableDefinitionV2) MembersAttrDefinition() *AttributeDefinition {
	return t.Attributes.Members
}

func (t TableDefinitionV2) AttrDefinitionMap() map[string]*AttributeDefinition {
	return map[string]*AttributeDefinition{
		PolicyAttrTypeResource: t.ResourceAttrDefinition(),
		PolicyAttrTypeActions:  t.ActionsAttrDefinition(),
		PolicyAttrTypeMembers:  t.MembersAttrDefinition(),
	}
}
