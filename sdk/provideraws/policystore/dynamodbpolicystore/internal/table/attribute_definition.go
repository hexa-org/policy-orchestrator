package table

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// AttributeDefinition both pk, sk cannot be true
type AttributeDefinition struct {
	//PolicyAttrType string // one of 'resource', 'actions' or 'members'
	NameOrPath string `json:"nameOrPath"`
	ValType    string `json:"valType"`
	Pk         bool   `json:"pk"`
	Sk         bool   `json:"sk"`
}

func NewAttributeDefinition(nameOrPath string, valType string, pk bool, sk bool) *AttributeDefinition {
	return &AttributeDefinition{NameOrPath: nameOrPath, ValType: valType, Pk: pk, Sk: sk}
}

func makeTableAttrName(fld reflect.StructField) string {
	jsonTag := fld.Tag.Get("json")
	jsonTagParts := strings.Split(jsonTag, ",")
	if len(jsonTagParts) == 0 {
		return ""
	}
	return jsonTagParts[0]
}

func validateAttrNameOrPathPart(nameOrPathPart string) error {
	tableAttrName := strings.TrimSpace(nameOrPathPart)
	// Allowed characters a-z, A-Z, 0-9, '_', '-' and '.'
	isMatch := regexp.MustCompile(`^[.A-Za-z0-9_-]*$`).MatchString(tableAttrName)
	if !isMatch || strings.HasPrefix(tableAttrName, ".") || strings.HasSuffix(tableAttrName, ".") {
		return fmt.Errorf("failed to validate nameOrPath. Only allowed [a-ZA-Z0-9_-.] (. not allowed as prefix or suffix). nameOrPathPart=%s", nameOrPathPart)
	}
	return nil
}

func makeValType(fld reflect.StructField) string {
	aValType := fld.Type.Name()
	if fld.Type.Kind() == reflect.Slice {
		aValType = "[]" + fld.Type.Elem().Name()
	}

	if aValType == "int" || aValType == "string" || aValType == "[]string" || aValType == "[]int" {
		return aValType
	}

	return ""
}

func validateValType(aValType string) error {
	if aValType == "int" || aValType == "string" || aValType == "[]string" || aValType == "[]int" {
		return nil
	}
	return fmt.Errorf("unsupported attribute type in definition. only allowed one of 'int','string','[]int','[]string'")
}

func makeMeta(fld reflect.StructField) (string, bool, bool) {
	metaTag := fld.Tag.Get("meta")
	metaParts := strings.Split(metaTag, ",")
	var isPk bool
	var isSk bool
	var attrNameIdentifier string

	for _, aPart := range metaParts {
		if aPart == PolicyAttrTypeResource || aPart == PolicyAttrTypeActions || aPart == PolicyAttrTypeMembers {
			attrNameIdentifier = aPart
		}

		if strings.Contains(strings.ReplaceAll(aPart, " ", ""), "pk") {
			isPk = true
		}
		if strings.Contains(strings.ReplaceAll(aPart, " ", ""), "sk") {
			isSk = true
		}
	}

	return attrNameIdentifier, isPk, isSk
}
