package dynamodbpolicystore

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
)

// TableDefinition - // Resource, Action, Members attribute names in the dynamodb table
// Clients provide this based on their dynamodb table definition
type TableDefinition struct {
	ResourceAttrName string
	ActionAttrName   string
	MembersAttrName  string
}

type TableInfo[R rar.ResourceActionRolesMapper] struct {
	TableName       string
	TableDefinition TableDefinition
	ItemType        R
}
