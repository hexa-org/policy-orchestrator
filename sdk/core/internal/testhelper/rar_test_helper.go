package testhelper

import "github.com/hexa-org/policy-orchestrator/sdk/core/rar"

// MakeRarMap - builds a map with as many elements as in the given httpMethods array
// key = "http:" + aMethod + resource
// value = rar.ResourceActionRoles
// e.g. { "http:GET/someresource": rar, "http:POST/someresource": rar }
func MakeRarMap(resource string, httpMethods []string, members []string) map[string]rar.ResourceActionRoles {
	return MakeRarMapMultiple([]string{resource}, [][]string{httpMethods}, [][]string{members})
}

// MakeRarMapMultiple builds a map with as many element as total len of httpMethods
// number of elements in each array param should be equal
// i.e. len(resources) == len(httpMethods) == len(members)
// individual member elements can be nil or empty
// e.g. MakeRarMapMultiple({"res1", "res2", {{GET,POST},{PUT}}, {{"mem1"}, {nil}})
// returns {
//    key1: rar(res1, GET, mem1,
//    key2: rar(res1, POST, mem1,
//    key3: rar(res2, PUT, nil
// }

func MakeRarMapMultiple(resources []string, httpMethods [][]string, members [][]string) map[string]rar.ResourceActionRoles {
	rarMap := make(map[string]rar.ResourceActionRoles)
	for i, aResource := range resources {
		for _, aMethod := range httpMethods[i] {
			lookupKey := makeRarKey(aResource, aMethod)
			aRar, _ := rar.NewResourceActionRoles(aResource, []string{aMethod}, members[i])
			rarMap[lookupKey] = aRar
		}
	}
	return rarMap
}

func MakeRarListMultiple(resources []string, httpMethods [][]string, members [][]string) []rar.ResourceActionRoles {
	rarList := make([]rar.ResourceActionRoles, 0)
	for i, aResource := range resources {
		for _, aMethod := range httpMethods[i] {
			aRar, _ := rar.NewResourceActionRoles(aResource, []string{aMethod}, members[i])
			rarList = append(rarList, aRar)
		}
	}
	return rarList
}

func makeRarKey(resource, method string) string {
	return method + resource
}
