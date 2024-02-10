package table

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	log "golang.org/x/exp/slog"
	"reflect"
	"strings"
)

type TableInfo[R rar.ResourceActionRolesMapper] struct {
	TableName          string
	ItemType           R
	ItemMappingDynamic bool // MapTo static, or dynamic. Decides whether to use item.MapTo() or MapTo(interface{})
	TableDefinition    TableDefinition
}

// NewTableInfo - creates a simple table info
// Only scalar attributes
// R is a struct with 3 elements having struct tags for 'meta'
// the tag identifies the element as a 'resource', 'actions', 'members'
// the type of element must be either scalar string, int or slice of string, int
// one of them must be pk
// max one can be pk, and one can be sk
// sk is optional
// should have one for 'members' and this cannot be a pk or sk
// attribute names can contain a-zA-Z0-9 or '_', '-' or '.'
//
//	'.' cannot be at the start or end of attribute name
/*
func NewTableInfo[R rar.ResourceActionRolesMapper](tableName string, sampleItem R) (*TableInfo[R], error) {

	attrMap := make(map[string]*AttributeDefinition)
	sType := reflect.TypeOf(sampleItem)
	for i := 0; i < sType.NumField(); i++ {
		fld := sType.Field(i)
		tableAttrName := makeTableAttrName(fld)
		theValType := makeValType(fld)
		policyAttrType, isPk, isSk := makeMeta(fld)
		aDef := &AttributeDefinition{
			NameOrPath: tableAttrName,
			ValType:    theValType,
			Pk:         isPk,
			Sk:         isSk,
		}
		attrMap[policyAttrType] = aDef
		log.Info("NewTableInfo", "tableAttrName", aDef.NameOrPath, "ValType", aDef.ValType, "PK", aDef.Pk, "SK", aDef.Sk)
	}

	tableDef := NewTableDefinition(attrMap)
	err := validateTableDefinition(false, tableDef)
	if err != nil {
		return nil, err
	}
	return &TableInfo[R]{TableName: tableName, ItemType: sampleItem, ItemMappingDynamic: false, TableDefinition: tableDef}, nil
}
*/

func NewSimpleTableInfo[R rar.ResourceActionRolesMapper](tableName string, sampleItem R) (*TableInfo[R], error) {
	tableDef, err := newSimpleTableDefinition(sampleItem)
	if err != nil {
		return nil, err
	}

	return &TableInfo[R]{TableName: tableName, ItemMappingDynamic: false, TableDefinition: tableDef}, nil
}

func NewDynamicTableInfo(tableName string, resourceAttrDef, actionsAttrDef, membersAttrDef *AttributeDefinition) (*TableInfo[rar.DynamicResourceActionRolesMapper], error) {
	tableDef, err := newDynamicTableDefinition(resourceAttrDef, actionsAttrDef, membersAttrDef)
	if err != nil {
		return nil, err
	}

	return &TableInfo[rar.DynamicResourceActionRolesMapper]{TableName: tableName, ItemMappingDynamic: true, TableDefinition: tableDef}, nil
}

func newSimpleTableDefinition[R rar.ResourceActionRolesMapper](sampleItem R) (TableDefinition, error) {
	var resourceAttrDef *AttributeDefinition
	var actionsAttrDef *AttributeDefinition
	var membersAttrDef *AttributeDefinition

	sType := reflect.TypeOf(sampleItem)
	for i := 0; i < sType.NumField(); i++ {
		fld := sType.Field(i)
		tableAttrName := makeTableAttrName(fld)
		theValType := makeValType(fld)
		policyAttrType, isPk, isSk := makeMeta(fld)
		aDef := &AttributeDefinition{
			NameOrPath: tableAttrName,
			ValType:    theValType,
			Pk:         isPk,
			Sk:         isSk,
		}

		log.Info("NewTableInfo", "tableAttrName", aDef.NameOrPath, "ValType", aDef.ValType, "PK", aDef.Pk, "SK", aDef.Sk)
		switch policyAttrType {
		case PolicyAttrTypeResource:
			resourceAttrDef = aDef
			break
		case PolicyAttrTypeActions:
			actionsAttrDef = aDef
			break
		case PolicyAttrTypeMembers:
			membersAttrDef = aDef
			break
		default:

		}
	}

	return newTableDefinition(resourceAttrDef, actionsAttrDef, membersAttrDef)

}

func newDynamicTableDefinition(resourceAttrDef, actionsAttrDef, membersAttrDef *AttributeDefinition) (TableDefinition, error) {
	return newTableDefinition(resourceAttrDef, actionsAttrDef, membersAttrDef)
}

func newTableDefinition(resourceAttrDef, actionsAttrDef, membersAttrDef *AttributeDefinition) (TableDefinition, error) {
	tableDef := TableDefinitionV2{
		Attributes: TableAttributes{
			Resource: resourceAttrDef,
			Actions:  actionsAttrDef,
			Members:  membersAttrDef,
		},
	}

	err := ValidateTableDefinition(tableDef)
	if err != nil {
		return nil, err
	}
	return tableDef, nil
}

/*
func NewDynamicTableInfo[R rar.ResourceActionRolesMapper](tableName string, tableDefinition TableDefinition) (*TableInfo[R], error) {

	err := validateTableDefinition(true, tableDefinition)
	if err != nil {
		return nil, err
	}
	return &TableInfo[R]{TableName: tableName, ItemMappingDynamic: true, TableDefinition: tableDefinition}, nil
}
*/

func ValidateTableDefinition(tableDef TableDefinition) error {

	pkNames := make([]string, 0)
	skNames := make([]string, 0)
	for policyAttrType, aDef := range tableDef.AttrDefinitionMap() {
		// Ensure all 3 are present i.e. resource, actions, members
		if aDef == nil {
			return fmt.Errorf("failed to validate table definition. Missing attribute definition for %s", policyAttrType)
		}

		nameOrPath := aDef.NameOrPath

		// Composite attributes only allowed on non-key
		// Composites not allowed if using simple table definition
		if strings.Contains(nameOrPath, "/") {
			if aDef.Pk || aDef.Sk {
				return fmt.Errorf("failed to validate table definition. pk, sk attribute cannot be composite. tableAttrName=%s", nameOrPath)
			}

			//if !itemMappingDynamic {
			//	return fmt.Errorf("failed to validate table definition. simple table definitions cannot define composite attributes. tableAttrName=%s", nameOrPath)
			//}
		}

		// validate each path part for bad characters
		for _, pathPart := range strings.Split(nameOrPath, "/") {
			err := validateAttrNameOrPathPart(pathPart)
			if err != nil {
				return err
			}
		}

		err := validateValType(aDef.ValType)
		if err != nil {
			return fmt.Errorf("nameOrPath %s: error %w", nameOrPath, err)
		}

		if aDef.Pk {
			pkNames = append(pkNames, aDef.NameOrPath)
		}
		if aDef.Sk {
			skNames = append(skNames, aDef.NameOrPath)
		}
	}

	// exactly one pk required
	if len(pkNames) != 1 {
		return fmt.Errorf("failed to build tableInfo, at least one attribute must be defined as pk")
	}

	// exactly 0 or 1 sk
	if len(skNames) > 1 {
		return fmt.Errorf("failed to build tableInfo, cannot have more than one attribute as sk")
	}

	// resource must be string or int
	resDef := tableDef.ResourceAttrDefinition()
	if resDef.ValType == "[]string" || resDef.ValType == "[]int" {
		return fmt.Errorf("resource attribute value cannot be slice")
	}

	// member cannot be either pk or sk
	membersDef := tableDef.MembersAttrDefinition()
	if membersDef.Pk || membersDef.Sk {
		return fmt.Errorf("failed to build tableInfo, invalid 'members' attr definition. 'members' MUST be defined and cannot be pk, sk")
	}

	return nil
}
