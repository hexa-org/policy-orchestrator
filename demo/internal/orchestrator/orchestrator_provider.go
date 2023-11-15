package orchestrator

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policyprovider"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	log "golang.org/x/exp/slog"
	"net/http"
)

type attributeDefinition struct {
	nameOrPath string
	valType    string
	pk         bool
	sk         bool
}

type tableDefinition struct {
	resource *attributeDefinition
	actions  *attributeDefinition
	members  *attributeDefinition
}

type TableDefinitionOpt func(t *tableDefinition)

func WithResourceAttrDefinition(nameOrPath string, valType string, pk bool, sk bool) TableDefinitionOpt {
	return func(t *tableDefinition) {
		t.resource = &attributeDefinition{
			nameOrPath: nameOrPath,
			valType:    valType,
			pk:         pk,
			sk:         sk,
		}
	}
}

func WithActionsAttrDefinition(nameOrPath string, valType string, pk bool, sk bool) TableDefinitionOpt {
	return func(t *tableDefinition) {
		t.actions = &attributeDefinition{
			nameOrPath: nameOrPath,
			valType:    valType,
			pk:         pk,
			sk:         sk,
		}
	}
}

func WithMembersAttrDefinition(nameOrPath string, valType string) TableDefinitionOpt {
	return func(t *tableDefinition) {
		t.members = &attributeDefinition{
			nameOrPath: nameOrPath,
			valType:    valType,
			pk:         false,
			sk:         false,
		}
	}
}

type OrchestrationProvider struct {
	service policyprovider.ProviderService
}

const awsPolicyStoreTableName = "ResourcePolicies"

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

const tableDefinitionV2Json = `
			{
				"metadata": {
					"pk": { "attribute": "resource" },
					"sk": { "attribute": "actions" }
				},
				"attributes": {
					"resource": { "nameOrPath": "Resource", "valType": "string", "pk": true },
					"actions": { "nameOrPath": "Action", "valType": "string", "sk": true },
					"members": { "nameOrPath": "Members", "valType": "string" }
				}  
			}`

func NewOrchestrationProviderWithDynamicTableInfo(idpCredentials []byte, policyStoreCredentials []byte, tableOpts ...TableDefinitionOpt) (*OrchestrationProvider, error) {

	log.Info("NewOrchestrationProviderWithDynamicTableInfo", "msg", "New")
	tableDef := &tableDefinition{}
	for _, aOpt := range tableOpts {
		aOpt(tableDef)
	}

	attrDef := tableDef.resource
	resDef := dynamodbpolicystore.NewAttributeDefinition(attrDef.nameOrPath, attrDef.valType, attrDef.pk, attrDef.sk)

	attrDef = tableDef.actions
	actionsDef := dynamodbpolicystore.NewAttributeDefinition(attrDef.nameOrPath, attrDef.valType, attrDef.pk, attrDef.sk)

	attrDef = tableDef.members
	membersDef := dynamodbpolicystore.NewAttributeDefinition(attrDef.nameOrPath, attrDef.valType, attrDef.pk, attrDef.sk)

	tableInfo, err := dynamodbpolicystore.NewDynamicTableInfo(awsPolicyStoreTableName, resDef, actionsDef, membersDef)
	policyStoreSvc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, policyStoreCredentials)
	if err != nil {
		log.Error("NewOrchestrationProviderWithDynamicTableInfo",
			"msg", "failed to create dynamodbpolicystore.PolicyStoreSvc",
			"error", err)
		return nil, err
	}

	appInfoSvc, err := cognitoidp.NewAppInfoSvc(idpCredentials)
	if err != nil {
		log.Error("NewAwsApiGatewayProviderV2",
			"msg", "failed to create cognitoidp.AppInfoSvc",
			"error", err)
		return nil, err
	}

	service := policyprovider.NewProviderService[resourcePolicyItem](appInfoSvc, policyStoreSvc)
	provider := &OrchestrationProvider{
		service: service,
	}
	return provider, nil
}
func NewOrchestrationProvider(idpCredentials []byte, policyStoreCredentials []byte) (*OrchestrationProvider, error) {
	tableInfo, err := dynamodbpolicystore.NewSimpleTableInfo(awsPolicyStoreTableName, resourcePolicyItem{})
	policyStoreSvc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, policyStoreCredentials)
	if err != nil {
		log.Error("NewOrchestrationProvider",
			"msg", "failed to create dynamodbpolicystore.PolicyStoreSvc",
			"error", err)
		return nil, err
	}

	appInfoSvc, err := cognitoidp.NewAppInfoSvc(idpCredentials)
	if err != nil {
		log.Error("NewAwsApiGatewayProviderV2",
			"msg", "failed to create cognitoidp.AppInfoSvc",
			"error", err)
		return nil, err
	}

	service := policyprovider.NewProviderService[resourcePolicyItem](appInfoSvc, policyStoreSvc)
	provider := &OrchestrationProvider{
		service: service,
	}
	return provider, nil
}

func (a *OrchestrationProvider) Name() string {
	return "amazon"
}

func (a *OrchestrationProvider) DiscoverApplications(integrationInfo IntegrationInfo) ([]ApplicationInfo, error) {
	apps, err := a.service.DiscoverApplications()
	if err != nil {
		return nil, err
	}

	retApps := make([]ApplicationInfo, 0)
	for _, oneApp := range apps {
		retApps = append(retApps, toApplicationInfo(oneApp))
	}

	return retApps, nil

}

func (a *OrchestrationProvider) GetPolicyInfo(info IntegrationInfo, applicationInfo ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	idpAppInfo := toIdpAppInfo(applicationInfo)
	return a.service.GetPolicyInfo(idpAppInfo)
}

func (a *OrchestrationProvider) SetPolicyInfo(info IntegrationInfo, applicationInfo ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (status int, foundErr error) {
	log.Info("SetPolicyInfo", "msg", "BEGIN",
		"applicationInfo.ObjectID", applicationInfo.ObjectID,
		"Name", applicationInfo.Name,
		"Description", applicationInfo.Description,
		"Service", applicationInfo.Service)

	idpAppInfo := toIdpAppInfo(applicationInfo)
	err := a.service.SetPolicyInfo(idpAppInfo, policyInfos)
	log.Info("SetPolicyInfo", "msg", "Finished calling service.SetPolicyInfo")

	if err != nil {
		log.Error("SetPolicyInfo", "msg", "error calling service.SetPolicyInfo", "error", err)
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func toApplicationInfo(anApp idp.AppInfo) ApplicationInfo {
	rsApp := (anApp).(cognitoidp.ResourceServerAppInfo)
	return ApplicationInfo{
		ObjectID:    rsApp.Id(),
		Name:        rsApp.Name(),
		Description: rsApp.DisplayName(),
		Service:     rsApp.Identifier(),
	}
}

func toIdpAppInfo(applicationInfo ApplicationInfo) idp.AppInfo {
	return cognitoidp.NewResourceServerAppInfo(applicationInfo.ObjectID, applicationInfo.Name, applicationInfo.Description, applicationInfo.Service)
}
