package orchestrator

import (
	"encoding/json"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policyprovider"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore"
	log "golang.org/x/exp/slog"
	"net/http"
)

type OrchestrationProvider struct {
	service policyprovider.ProviderService
}

const awsPolicyStoreTableName = "ResourcePolicies"

var tableDefinition = dynamodbpolicystore.TableDefinition{
	ResourceAttrName: "ResourceX",
	ActionAttrName:   "ActionX",
	MembersAttrName:  "MembersX",
}

type resourcePolicyItem struct {
	Resource string `json:"Resource"`
	Action   string `json:"Action"`
	Members  string `json:"Members"`
}

func (t resourcePolicyItem) MapTo() (rar.ResourceActionRoles, error) {
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

/*var tableDefinitionV2 = dynamodbpolicystore.TableDefinitionV2{
	Metadata: struct {
		Pk dynamodbpolicystore.MetadataKeyInfo `json:"pk"`
		Sk dynamodbpolicystore.MetadataKeyInfo `json:"sk"`
	}{
		Pk: dynamodbpolicystore.MetadataKeyInfo{Attribute: "resource"},
		Sk: dynamodbpolicystore.MetadataKeyInfo{Attribute: "actions"},
	},
	Attributes: struct {
		Resource dynamodbpolicystore.AttributeDefinition `json:"resource"`
		Actions  dynamodbpolicystore.AttributeDefinition `json:"actions"`
		Members  dynamodbpolicystore.AttributeDefinition `json:"members"`
	}{
		Resource: dynamodbpolicystore.AttributeDefinition{
			NameOrPath: "ResourceX",
			ValType:    "string",
		},
		Actions: dynamodbpolicystore.AttributeDefinition{
			NameOrPath: "ActionsX",
			ValType:    "string",
		},
		Members: dynamodbpolicystore.AttributeDefinition{
			NameOrPath: "MembersX",
			ValType:    "string",
		},
	},
}*/

const tableDefinitionV2Json2 = `
			{
				"metadata": {
					"pk": { "attribute": "resource" },
					"sk": { "attribute": "actions" }
				},
				"attributes": {
					"resource": { "nameOrPath": "Policy/Nested/ResourceX", "valType": "string" },
					"actions": { "nameOrPath": "Policy/Nested/ActionsX", "valType": "string[]" },
					"members": { "nameOrPath": "Policy/Nested/Members", "valType": "int[]" }
				}  
			}`

const tableDefinitionV2Json = `
			{
				"metadata": {
					"pk": { "attribute": "resource" },
					"sk": { "attribute": "actions" }
				},
				"attributes": {
					"resource": { "nameOrPath": "Resource", "valType": "string" },
					"actions": { "nameOrPath": "Action", "valType": "string" },
					"members": { "nameOrPath": "Members", "valType": "string" }
				}  
			}`

func (it flexibleItem) MapToV2(scanOutputItem interface{}) (rar.ResourceActionRoles, error) {
	log.Info("MapToV2", "scanOutputItem", scanOutputItem)
	theMap := scanOutputItem.(map[string]interface{})

	aRes := theMap["Resource"].(string)
	anAct := theMap["Action"].(string)
	aMemStr := theMap["Members"].(string)

	members := make([]string, 0)
	_ = json.Unmarshal([]byte(aMemStr), &members)
	return rar.NewResourceActionRoles(aRes, []string{anAct}, members)
}

func NewOrchestrationProvider(idpCredentials []byte, policyStoreCredentials []byte) (*OrchestrationProvider, error) {
	/*tableInfo := dynamodbpolicystore.TableInfo[resourcePolicyItem]{
		TableName:       awsPolicyStoreTableName,
		TableDefinition: tableDefinition,
		ItemType:        resourcePolicyItem{},
	}*/

	var tableDefinitionV2 dynamodbpolicystore.TableDefinitionV2
	err := json.Unmarshal([]byte(tableDefinitionV2Json), &tableDefinitionV2)
	if err != nil {
		log.Error("NewOrchestrationProvider", "msg", "failed to marshall tableDefinition json", "error", err)
		return nil, err
	}

	tableInfo := dynamodbpolicystore.TableInfo[flexibleItem]{
		TableName:          awsPolicyStoreTableName,
		TableDefinition:    tableDefinition,
		ItemType:           flexibleItem{},
		ItemMappingDynamic: true,
		TableDefinitionV2:  tableDefinitionV2,
	}

	appInfoSvc, err := cognitoidp.NewAppInfoSvc(idpCredentials)
	if err != nil {
		log.Error("NewAwsApiGatewayProviderV2",
			"msg", "failed to create cognitoidp.AppInfoSvc",
			"error", err)
		return nil, err
	}

	policyStoreSvc, err := dynamodbpolicystore.NewPolicyStoreSvc(tableInfo, policyStoreCredentials)
	if err != nil {
		log.Error("NewAwsApiGatewayProviderV2",
			"msg", "failed to create dynamodbpolicystore.PolicyStoreSvc",
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
