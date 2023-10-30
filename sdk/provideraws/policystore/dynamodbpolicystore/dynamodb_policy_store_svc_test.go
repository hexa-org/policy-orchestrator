package dynamodbpolicystore_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func TestNewPolicyStoreSvc_Error(t *testing.T) {
	tableInfo := dynamodbpolicystore.TableInfo[testhelper.TestTableItem]{}
	svc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, []byte("$$"))
	assert.ErrorContains(t, err, "invalid character '$'")
	assert.Nil(t, svc)

}

func TestNewPolicyStoreSvc(t *testing.T) {
	tableInfo := dynamodbpolicystore.TableInfo[testhelper.TestTableItem]{
		TableName:       testhelper.TableName,
		TableDefinition: testhelper.TableDefinition(),
		ItemType:        testhelper.TestTableItem{},
	}
	svc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, testhelper.AwsCredentialsForTest())

	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestGetPolicies_ScanError(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	app := new(idp.AppInfo)
	c.ExpectScan(errors.New("some-error"))
	policies, err := svc.GetPolicies(*app)
	assert.ErrorContains(t, err, "some-error")
	assert.Nil(t, policies)
}

func TestGetPolicies_EmptyItemsFromScan(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	app := new(idp.AppInfo)
	c.ExpectScan(nil)
	policies, err := svc.GetPolicies(*app)
	assert.NoError(t, err)
	assert.NotNil(t, policies)
	assert.Empty(t, policies)
}
func TestGetPolicies(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	app := new(idp.AppInfo)
	expRar := testhelper.MakeResourceActionRoles()

	c.ExpectScan(nil, expRar)
	policies, err := svc.GetPolicies(*app)
	assert.NoError(t, err)
	assert.NotNil(t, policies)
	assert.Equal(t, []rar.ResourceActionRoles{expRar}, policies)
}

func TestSetPolicy(t *testing.T) {
	svc, c := newPolicyStoreSvcAndClient()
	expRar := testhelper.MakeResourceActionRoles()
	c.ExpectUpdateItem(expRar, nil)
	err := svc.SetPolicy(expRar)
	assert.NoError(t, err)
}

type flexibleItem struct {
}

func (it flexibleItem) MapTo() (rar.ResourceActionRoles, error) {
	/*res := it.Fields.(map[string]interface{})["ResourceX"].(string)
	actions := it.Fields.(map[string]interface{})["ActionX"].(string)
	memStr := it.Fields.(map[string]interface{})["MembersX"].(string)

	members := make([]string, 0)
	_ = json.Unmarshal([]byte(memStr), &members)
	return rar.NewResourceActionRoles(res, []string{actions}, members)*/
	panic("MapTo() is deprecated")
}

func (it flexibleItem) MapToV2(scanOutputItem interface{}) (rar.ResourceActionRoles, error) {
	theMap := scanOutputItem.(map[string]interface{})

	aRes := theMap["ResourceX"].(string)
	anAct := theMap["ActionX"].(string)
	aMemStr := theMap["MembersX"].(string)

	members := make([]string, 0)
	_ = json.Unmarshal([]byte(aMemStr), &members)
	return rar.NewResourceActionRoles(aRes, []string{anAct}, members)
}

func TableDefinition() dynamodbpolicystore.TableDefinition {
	return dynamodbpolicystore.TableDefinition{
		ResourceAttrName: "ResourceX",
		ActionAttrName:   "ActionX",
		MembersAttrName:  "MembersX",
	}
}

func TestWithDynamicItemJson(t *testing.T) {
	tableItemInfo := flexibleItem{}
	tableInfo := dynamodbpolicystore.TableInfo[flexibleItem]{
		TableName:          testhelper.TableName,
		TableDefinition:    testhelper.TableDefinition(),
		ItemType:           tableItemInfo,
		ItemMappingDynamic: true,
	}

	c := testhelper.NewMockClient(testhelper.TableDefinition())
	opt := dynamodbpolicystore.WithDynamodbClientOverride[flexibleItem](c)
	svc, _ := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, testhelper.AwsCredentialsForTest(), opt)
	app := new(idp.AppInfo)
	expRar := testhelper.MakeResourceActionRoles()

	c.ExpectScan(nil, expRar)
	policies, err := svc.GetPolicies(*app)
	assert.NoError(t, err)
	assert.NotNil(t, policies)
	assert.Equal(t, []rar.ResourceActionRoles{expRar}, policies)
}

func TestItemInterface(t *testing.T) {
	type someItem interface {
	}
	//str := `{ "Resource": "A-Resource", "Action": "An-Action", "Members": "some member", "Nested": { "Resource": "Child-Resource" }	}`

	items := make([]map[string]someItem, 0)
	item := make(map[string]someItem)
	item["ResourceX"] = ddbTypeString("A-Resource")
	item["ActionX"] = ddbTypeString("An-Action") // TODO handle array
	members := fmt.Sprintf("[\"%s\"]", "some member")
	item["MembersX"] = ddbTypeString(members)
	items = append(items, item)
	fmt.Println(items)
}

type dynamodbAttributeAccessors struct {
	AttrName  string
	ValueType string
	//AttrValue        interface{}
}

func createSample(aMap map[string]dynamodbAttributeAccessors) map[string]interface{} {

	nestedObject := make(map[string]interface{})
	fieldsToIterate := []string{"resource", "actions", "members"}
	for _, aField := range fieldsToIterate {
		attrDef := aMap[aField]
		attrPathParts := strings.Split(attrDef.AttrName, ".")

		curr := nestedObject
		for _, part := range attrPathParts {
			_, found := curr[part]
			if !found {
				newMap := make(map[string]interface{})
				curr[part] = newMap
				curr = newMap
			} else {
				curr = curr[part].(map[string]interface{})
			}
		}
	}

	return nestedObject

}

func parseDynamodbAttributeAccessors() (map[string]dynamodbAttributeAccessors, error) {
	var aMap map[string]dynamodbAttributeAccessors

	// pk: attrValue can only be a scalar string, or int
	// sk: attrValue can only be a scalar string, or int
	// resource: string or int
	// actions: string or int or array of string or int
	// members: string or int or array of string or int

	jsonStr := `{
	 "pk": {
		"attrName": "ResourceX", "attrValue": 1, "valueType": "int"
	 }, 
	 "sk": {
		"attrName": "ActionX", "attrValue": 1, "valueType": "int"
	 }, 
	  "resource": {
		"attrName": "Policy.Nested.ResourceX", "attrValue": 1, "valueType": "int"
      },
	  "actions": {
		"attrName": "Policy.Nested.ActionX", "attrValue": "example", "valueType": "string"
      },
	  "members": {
		"attrName": "Policy.Nested.MembersX", "attrValue": "[\"mem1\", \"mem2\"]", "valueType": "[]string"
	  }
	}`

	err := json.Unmarshal([]byte(jsonStr), &aMap)

	return aMap, err
}

func TestAttributeAccessors(t *testing.T) {
	aMap, err := parseDynamodbAttributeAccessors()
	fmt.Println(err)
	fmt.Println(aMap)

	nestedObject := createSample(aMap)
	fmt.Println("nestedObject", nestedObject)

	/*type someItem interface {
	}
	items := make([]map[string]someItem, 0)
	item := make(map[string]someItem)
	item["ResourceX"] = ddbTypeString("A-Resource")
	item["ActionX"] = ddbTypeString("An-Action") // TODO handle array
	members := fmt.Sprintf("[\"%s\"]", "some member")
	item["MembersX"] = ddbTypeString(members)
	items = append(items, item)
	fmt.Println(items)

	outputItems := make([]map[string]types.AttributeValue, 0)
	oneItem := make(map[string]types.AttributeValue)

	//resourceAttrDef := aMap["resource"]

	oneItem["ResourceX"] = &types.AttributeValueMemberS{Value: "aResource"}

	numArr := make([]types.AttributeValue, 0)
	numArr = append(numArr, &types.AttributeValueMemberN{Value: "1"})
	numArr = append(numArr, &types.AttributeValueMemberN{Value: "2"})
	oneItem["ActionX"] = &types.AttributeValueMemberL{Value: numArr}

	memArr := make([]types.AttributeValue, 0)
	memArr = append(numArr, &types.AttributeValueMemberS{Value: "mem1"})
	memArr = append(numArr, &types.AttributeValueMemberS{Value: "mem2"})
	oneItem["MembersX"] = &types.AttributeValueMemberL{Value: memArr}

	outputItems = append(outputItems, oneItem)
	//scanOutput := ddb.ScanOutput{Items: outputItems}*/
}

// currItem := mapItem
//
//	Policy: {
//		Nested: {
//		  ResourceX: aResource
//	 }
//	}
func TestNestedMapResource(t *testing.T) {
	//oneItem := make(map[string]types.AttributeValue)
	aRar, _ := rar.NewResourceActionRoles("12", []string{"GET"}, []string{"mem1", "mem2"})

	aMap, err := parseDynamodbAttributeAccessors()
	fmt.Println(err)
	fmt.Println(aMap)

	oneItem := make(map[string]types.AttributeValue)

	resourceAttrDef := aMap["resource"]                                               //"Policy.Nested.ResourceX"
	var mainAttrKey, mainItem = toAttributeValueMap(resourceAttrDef, aRar.Resource()) // Policy value
	oneItem[mainAttrKey] = mainItem

	resourceAttrDef = aMap["actions"]                                            //"Policy.Nested.ActionX"
	mainAttrKey, mainItem = toAttributeValueMap(resourceAttrDef, aRar.Actions()) // Policy value
	oneItem[mainAttrKey] = mainItem

	resourceAttrDef = aMap["members"]                                            //"Policy.Nested.Members" is an array
	mainAttrKey, mainItem = toAttributeValueMap(resourceAttrDef, aRar.Members()) // Policy value
	oneItem[mainAttrKey] = mainItem

	fmt.Println(oneItem)
	str, err := json.Marshal(oneItem)
	fmt.Println(err)
	fmt.Println(string(str))
}

func toAttributeValueMap(accessor dynamodbAttributeAccessors, attrValue interface{}) (mainItemAttrName string, mainItem *types.AttributeValueMemberM) {
	if !strings.Contains(accessor.AttrName, ".") {
		fmt.Println("toAttributeValueMap - not a map type accessor, does not contain '.'")
		return "", nil
	}
	objKeys := strings.Split(accessor.AttrName, ".")
	mainItemAttrName = objKeys[0]

	// its a map
	mainItem = &types.AttributeValueMemberM{} // Policy value
	currItem := mainItem
	currItem.Value = map[string]types.AttributeValue{}

	nextItem := &types.AttributeValueMemberM{} // Nested
	currItem.Value[objKeys[1]] = nextItem
	nextItem.Value = map[string]types.AttributeValue{}

	// for int, this returns float64. Somehow the json.Marshall uses float64 as the underlying type
	//attributeKind := reflect.TypeOf(accessor.AttrValue).Kind()
	valueType := accessor.ValueType
	var lastValue types.AttributeValue

	if valueType == "[]string" || valueType == "[]int" {
		//elemKind := reflect.TypeOf(accessor.AttrValue).Elem().Kind()
		arrValues := make([]types.AttributeValue, 0)
		useValues, ok := attrValue.([]string)
		if !ok {
			fmt.Println("toAttributeValueMap - attribute value defined as slice, but could not convert to []string",
				"accessor.AttrName", accessor.AttrName, "attrValue", attrValue)
			return "", nil
		}
		if valueType == "[]int" {
			for _, strVal := range useValues {
				_, err := strconv.Atoi(strVal)
				if err != nil {
					fmt.Println("toAttributeValueMap - attribute defined as []int, but could not convert value to int.",
						"accessor.AttrName", accessor.AttrName, "attrValue", attrValue)
					return "", nil
				}
				arrValues = append(arrValues, &types.AttributeValueMemberN{Value: strVal})
			}
		} else if valueType == "[]string" {
			for _, strVal := range useValues {
				arrValues = append(arrValues, &types.AttributeValueMemberS{Value: strVal})
			}
		} else {
			fmt.Println("toAttributeValueMap - if slice, value needs to be a string or int",
				"accessor.AttrName", accessor.AttrName, "attrValue", attrValue)
			return "", nil
		}

		lastValue = &types.AttributeValueMemberL{Value: arrValues}
	} else {
		// rar uses arrays to hold even scalar values e.g. Actions
		// in this case we expect an slice of size <= 1
		// if len == 0, we will use an empty AttributeValue e.g. if removing all
		useValues, ok := attrValue.([]string)

		if ok && len(useValues) > 1 {
			fmt.Println("toAttributeValueMap - attribute defined as scalar, but value found to be slice with len > 1",
				"accessor.AttrName", accessor.AttrName, "attrValue", attrValue)
			return "", nil
		}

		var strVal string
		if ok && len(useValues) == 1 {
			strVal = useValues[0]
		} else {
			strVal, ok = attrValue.(string)
			if !ok {
				fmt.Println("toAttributeValueMap - attribute defined as scalar, but could not convert to string.",
					"accessor.AttrName", accessor.AttrName, "attrValue", attrValue)
				return "", nil
			}
		}

		if valueType == "string" {
			lastValue = &types.AttributeValueMemberS{Value: strVal}
		} else if valueType == "int" {
			_, err := strconv.Atoi(strVal)
			if err != nil {
				fmt.Println("toAttributeValueMap - attribute defined as int, but could not convert value to int.",
					"accessor.AttrName", accessor.AttrName, "attrValue", attrValue)
				return "", nil
			}
			lastValue = &types.AttributeValueMemberS{Value: strVal}
		}
	}

	nextItem.Value[objKeys[2]] = lastValue
	return
}

type scanOutputType struct {
	Items []map[string]interface{}
}

func ddbTypeString(val string) map[string]interface{} {
	return map[string]interface{}{
		"S": val,
	}
}

func newPolicyStoreSvcAndClient() (policystore.PolicyBackendSvc[testhelper.TestTableItem], *testhelper.MockClient) {
	tableItemInfo := testhelper.TestTableItem{}
	tableInfo := dynamodbpolicystore.TableInfo[testhelper.TestTableItem]{
		TableName:       testhelper.TableName,
		TableDefinition: testhelper.TableDefinition(),
		ItemType:        tableItemInfo,
	}

	mockClient := testhelper.NewMockClient(testhelper.TableDefinition())
	opt := dynamodbpolicystore.WithDynamodbClientOverride[testhelper.TestTableItem](mockClient)
	svc, _ := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, testhelper.AwsCredentialsForTest(), opt)
	return svc, mockClient
}
