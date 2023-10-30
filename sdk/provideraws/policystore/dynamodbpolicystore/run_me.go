package dynamodbpolicystore

import (
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
)

// resourcePolicyItem
// Example begin
type resourcePolicyItem struct {
	Resource string `json:"Resource"`
	Action   string `json:"Action"`
	Members  string `json:"Members"`
}

func (it *resourcePolicyItem) MapTo() (rar.ResourceActionRoles, error) {
	members := make([]string, 0)
	_ = json.Unmarshal([]byte(it.Members), &members)
	return rar.NewResourceActionRoles(it.Resource, []string{it.Action}, members)
}

func (it *resourcePolicyItem) MapToV2(item interface{}) (rar.ResourceActionRoles, error) {
	panic("Implement me in runMe MapToV2")
}

func RunMe() {
	tableName := "AwsPolicyStoreTableName"
	dbInfo := TableInfo[*resourcePolicyItem]{
		TableName: tableName,
		TableDefinition: TableDefinition{
			ResourceAttrName: "Resource",
			ActionAttrName:   "Action",
			MembersAttrName:  "Members",
		},
		ItemType: &resourcePolicyItem{},
	}
	svc, err := NewPolicyStoreSvc(dbInfo, []byte("key"))
	app := new(idp.AppInfo)
	_, err = svc.GetPolicies(*app)
	if err != nil {
		return
	}
}

// Example end
