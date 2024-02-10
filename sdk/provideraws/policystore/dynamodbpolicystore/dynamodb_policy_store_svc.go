package dynamodbpolicystore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/client"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/table"
	log "golang.org/x/exp/slog"
	"sort"
	"strings"
)

type PolicyStoreSvc[R rar.ResourceActionRolesMapper] struct {
	client    client.DynamodbClient
	tableInfo *table.TableInfo[R]
}

type Opt[R rar.ResourceActionRolesMapper] func(svc *PolicyStoreSvc[R])

func WithDynamodbClientOverride[R rar.ResourceActionRolesMapper](client client.DynamodbClient) Opt[R] {
	return func(svc *PolicyStoreSvc[R]) {
		svc.client = client
	}
}

/*
func NewPolicyStoreSvcSimpleTable[R rar.ResourceActionRolesMapper](tableName string, tableItem R, key []byte, opts ...Opt[R]) (policystore.PolicyBackendSvc[R], error) {
	aTableInfo, err := table.NewTableInfo[R](tableName, tableItem)
	if err != nil {
		return nil, err
	}

	return newPolicyStoreSvc(aTableInfo, key, opts...)
}

func NewPolicyStoreSvcDynamicTable(tableName string, tableDef table.TableDefinition, key []byte, opts ...Opt[rar.DynamicResourceActionRolesMapper]) (policystore.PolicyBackendSvc[rar.DynamicResourceActionRolesMapper], error) {
	aTableInfo, err := table.NewDynamicTableInfo[rar.DynamicResourceActionRolesMapper](tableName, tableDef)
	//aTableInfo, err := NewTableInfo[rar.DynamicResourceActionRolesMapper](tableName, rar.DynamicResourceActionRolesMapper{})
	if err != nil {
		return nil, err
	}

	return newPolicyStoreSvc[rar.DynamicResourceActionRolesMapper](aTableInfo, key, opts...)
}
*/

func newPolicyStoreSvc[R rar.ResourceActionRolesMapper](tableInfo *table.TableInfo[R], key []byte, opts ...Opt[R]) (policystore.PolicyBackendSvc[R], error) {
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

func NewPolicyStoreSvc[R rar.ResourceActionRolesMapper](tableInfo *table.TableInfo[R], key []byte, opts ...Opt[R]) (policystore.PolicyBackendSvc[R], error) {
	if tableInfo == nil {
		return nil, errors.New("failed to create PolicyStoreSvc without tableInfo")
	}

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

	var rarList []rar.ResourceActionRoles

	if s.tableInfo.ItemMappingDynamic {
		rarList, err = s.getPoliciesDynamic(output)
	} else {
		// If can be mapped to a provided struct, then use simple MapTo
		rarList, err = s.getPoliciesSimple(output)
	}

	if err != nil {
		return nil, err
	}

	sortResourceActionRoleList(rarList)
	return rarList, nil
}

func (s *PolicyStoreSvc[R]) getPoliciesDynamic(output *ddb.ScanOutput) ([]rar.ResourceActionRoles, error) {
	var items []interface{}
	err := attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to unmarshal items. Err=", err)
		return nil, err
	}
	return s.toRarListV2(items)
}

func (s *PolicyStoreSvc[R]) getPoliciesSimple(output *ddb.ScanOutput) ([]rar.ResourceActionRoles, error) {
	var items []R
	err := attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to unmarshal items. Err=", err)
		return nil, err
	}

	return simpleRarListMapper(items)
}

func simpleRarListMapper[R rar.ResourceActionRolesMapper](items []R) ([]rar.ResourceActionRoles, error) {
	rarList := make([]rar.ResourceActionRoles, 0)
	for _, item := range items {
		aRar, err := item.MapTo()
		if err != nil {
			log.Error("ToResourceActionRoleList", "failed to map item to ResourceActionRoles", err)
			return nil, err
		}
		rarList = append(rarList, aRar)
	}

	return rarList, nil
}

func sortResourceActionRoleList(rarList []rar.ResourceActionRoles) {
	sort.SliceStable(rarList, func(i, j int) bool {
		resComp := strings.Compare(rarList[i].Resource(), rarList[j].Resource())
		actComp := strings.Compare(rarList[i].Actions()[0], rarList[j].Actions()[0]) // TODO handle array
		switch resComp {
		case 0:
			return actComp <= 0
		default:
			return resComp < 0
		}
	})
}

func (s *PolicyStoreSvc[R]) SetPolicy(aRar rar.ResourceActionRoles) error {
	log.Info("PolicyStoreSvc.SetPolicy", "msg", "aRar", aRar)
	builderV2 := client.NewInputBuilderV2(s.tableInfo.TableName, s.tableInfo.TableDefinition)

	input, err := builderV2.UpdateItemInput(aRar)
	if err != nil {
		log.Error("PolicyStoreSvc.SetPolicy", "msg", "failed to build updateItemInput", "error", err)
		return err
	}

	// TODO - process output
	log.Error("PolicyStoreSvc.SetPolicy", "msg", input)
	_, err = s.client.UpdateItem(context.TODO(), input)
	return err
}

func (s *PolicyStoreSvc[R]) toRarListV2(scanOutputItems []interface{}) ([]rar.ResourceActionRoles, error) {

	rarList := make([]rar.ResourceActionRoles, 0)

	for _, anItem := range scanOutputItems {
		log.Info("toRarListV2", "anItem", anItem)
		theMap := anItem.(map[string]interface{})
		aRar, err := getRarFromItem(s.tableInfo.TableDefinition, theMap)
		if err != nil {
			return nil, err
		}
		rarList = append(rarList, aRar)

	}
	return rarList, nil
}

func getRarFromItem(tableDef table.TableDefinition, theMap map[string]interface{}) (rar.ResourceActionRoles, error) {
	var resource string
	var actions []string
	var members []string

	for policyAttrType, aDef := range tableDef.AttrDefinitionMap() {
		valArr, err := parseItemValueForRar(aDef, theMap)
		if err != nil {
			return rar.ResourceActionRoles{}, err
		}

		if len(valArr) == 0 {
			continue
		}

		switch policyAttrType {
		case table.PolicyAttrTypeResource:
			resource = valArr[0]
			break
		case table.PolicyAttrTypeActions:
			actions = valArr
			break
		case table.PolicyAttrTypeMembers:
			members = valArr
			break
		default:
			return rar.ResourceActionRoles{}, fmt.Errorf("invalid PolicyAttrType %s. Should be one of [resource, actions, members]. AttrNameOrPath=%s", policyAttrType, aDef.NameOrPath)
		}
	}

	return rar.NewResourceActionRoles(resource, actions, members)
}

// parseValue returns []string irrespective or attribute value type
// caller to decide whether to use array fully, or just first element
// while building rar
func parseItemValueForRar(aDef *table.AttributeDefinition, theMap map[string]interface{}) ([]string, error) {
	nameOrPath := aDef.NameOrPath
	var aVal interface{}
	tmpMap := theMap
	if aDef.Pk || aDef.Sk {
		aVal = theMap[nameOrPath]
	} else {
		// non key - can be composite
		// value can be []string, []int, string, int
		// just get the leaf value here. dont worry about value type.
		partsArr := strings.Split(nameOrPath, "/")
		numParts := len(partsArr)

		// last part is not a map so only loop till 2nd last
		for _, aPart := range partsArr[0 : numParts-1] {
			tmpMap = tmpMap[aPart].(map[string]interface{})
		}

		attrName := partsArr[numParts-1] // last one is the actual attribute
		aVal = tmpMap[attrName]
	}

	var strArr []string
	if aDef.ValType == "[]string" {
		strArr = aVal.([]string)
	} else if aDef.ValType == "[]int" {
		arrVal := aVal.([]int)
		strArr = make([]string, 0)
		for _, v := range arrVal {
			strArr = append(strArr, fmt.Sprintf("%v", v))
		}
	} else if aDef.ValType == "string" {
		// takes care of int, string
		strVal := fmt.Sprintf("%s", aVal)
		strVal = strings.TrimSpace(strVal)
		// Check if its an array encoded as a string
		// e.g. members = "[\"Read.HRUS\"]"
		// in our test dynamodb instance, we use a string type for members, but
		// it actually supports multiple members "[\"Read.HRUS\", \"Read.HRUK\"]"
		if strings.HasPrefix(strVal, "[") && strings.HasSuffix(strVal, "]") {
			err := json.Unmarshal([]byte(strVal), &strArr)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshall string as array for %s. string value %s: error %w", nameOrPath, strVal, err)
			}
		} else {
			strArr = []string{fmt.Sprintf("%s", aVal)}
		}
	} else {
		// its an int
		strArr = []string{fmt.Sprintf("%v", aVal)}
	}

	return strArr, nil
}

/*
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

*/
