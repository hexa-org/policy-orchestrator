package providersV2

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	log "golang.org/x/exp/slog"
)

const AwsPolicyStoreTableName = "ResourcePolicies"

// resourcePolicyItem - definition of item stored in dynamodb
// meta tags are required
// This item specifies
// "Resource" column of type string, as the primary key
// "Action" column of type string, as the sort key
// "Members" column of type string, non-key
type resourcePolicyItem struct {
	Resource string `json:"Resource" meta:"resource,pk"`
	Action   string `json:"Action" meta:"actions,sk"`
	Members  string `json:"Members" meta:"members"`
}

func (t resourcePolicyItem) MapTo() (rar.ResourceActionRoles, error) {
	log.Info("resourcePolicyItem.MapTo", "msg", "Mapping", "rar", fmt.Sprintf("%v", t))
	members := make([]string, 0)
	err := json.Unmarshal([]byte(t.Members), &members)
	if err != nil {
		log.Error("resourcePolicyItem.MapTo", "msg", "Failed to unmarshal members string",
			"members", t.Members,
			"Err", err)
		return rar.ResourceActionRoles{}, err
	}
	return rar.NewResourceActionRoles(t.Resource, []string{t.Action}, members)
}

type simpleItemStore struct {
	tableName  string
	key        []byte
	simpleItem resourcePolicyItem
}

func NewSimpleItemStore(tableName string, key []byte) PolicyStore[rar.ResourceActionRolesMapper] {
	item := resourcePolicyItem{}
	return &simpleItemStore{tableName: tableName, key: key, simpleItem: item}
}

func (dps *simpleItemStore) Provider() (policystore.PolicyBackendSvc[rar.ResourceActionRolesMapper], error) {
	tableInfo, err := dynamodbpolicystore.NewSimpleTableInfo(dps.tableName, dps.simpleItem)
	policyStoreSvc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, dps.key)
	if err != nil {
		log.Error("NewOrchestrationProvider",
			"msg", "failed to create dynamodbpolicystore.PolicyStoreSvc",
			"error", err)
		return nil, err
	}
	return policyStoreSvc, nil
}

type dynamicItemStore struct {
	tableName string
	key       []byte
	tableDef  *tableDefinition
}

func NewDynamicItemStore(tableName string, key []byte, tableDef *tableDefinition) PolicyStore[rar.DynamicResourceActionRolesMapper] {
	return &dynamicItemStore{tableName: tableName, key: key, tableDef: tableDef}
}

func (dpd *dynamicItemStore) Provider() (policystore.PolicyBackendSvc[rar.DynamicResourceActionRolesMapper], error) {
	log.Info("NewOrchestrationProviderWithDynamicTableInfo", "msg", "New")

	attrDef := dpd.tableDef.resource
	resDef := dynamodbpolicystore.NewAttributeDefinition(attrDef.nameOrPath, attrDef.valType, attrDef.pk, attrDef.sk)

	attrDef = dpd.tableDef.actions
	actionsDef := dynamodbpolicystore.NewAttributeDefinition(attrDef.nameOrPath, attrDef.valType, attrDef.pk, attrDef.sk)

	attrDef = dpd.tableDef.members
	membersDef := dynamodbpolicystore.NewAttributeDefinition(attrDef.nameOrPath, attrDef.valType, attrDef.pk, attrDef.sk)

	tableInfo, err := dynamodbpolicystore.NewDynamicTableInfo(dpd.tableName, resDef, actionsDef, membersDef)
	policyStoreSvc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, dpd.key)
	if err != nil {
		log.Error("NewOrchestrationProviderWithDynamicTableInfo",
			"msg", "failed to create dynamodbpolicystore.PolicyStoreSvc",
			"error", err)
		return nil, err
	}
	return policyStoreSvc, nil
}

type attributeDefinition struct {
	nameOrPath string
	valType    string
	pk         bool
	sk         bool
}

func NewAttributeDefinition(nameOrPath string, valType string, pk bool, sk bool) *attributeDefinition {
	return &attributeDefinition{nameOrPath: nameOrPath, valType: valType, pk: pk, sk: sk}
}

type tableDefinition struct {
	resource *attributeDefinition
	actions  *attributeDefinition
	members  *attributeDefinition
}

func NewTableDefinition(resource *attributeDefinition, actions *attributeDefinition, members *attributeDefinition) *tableDefinition {
	return &tableDefinition{resource: resource, actions: actions, members: members}
}
