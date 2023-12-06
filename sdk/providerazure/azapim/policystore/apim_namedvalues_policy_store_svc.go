package policystore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/clientsupport"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/policystore/internal/client"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azurecommon"
	logger "golang.org/x/exp/slog"
	"net/http"
	"strings"
	"time"
)

const nvKeySep = "-"
const rarNVPrefix = "resrol"
const providerKeyActionPrefix = "http"

type NamedValuePolicyStoreSvc struct {
	client        client.NamedValuesClient
	resourceGroup string
	serviceName   string
}

func NewNamedValuePolicyStoreSvc(key []byte, httpClient azurecommon.HTTPClient, resourceGroup string, apimServiceName string) (policystore.PolicyBackendSvc[any], error) {
	client, err := client.NewNamedValuesClient(key, httpClient)
	if err != nil {
		return nil, err
	}
	return &NamedValuePolicyStoreSvc{client: client, resourceGroup: resourceGroup, serviceName: apimServiceName}, nil
}

func (n *NamedValuePolicyStoreSvc) GetPolicies(_ idp.AppInfo) ([]rar.ResourceActionRoles, error) {
	//s, ok := info.(*model.ArmApiAppInfo)

	//if !ok {
	//	return nil, fmt.Errorf("failed to get policies, expecting app of type model.ArmApiAppInfo")
	//}

	//logger.Info("NamedValuePolicyStoreSvc.GetPolicies", "service", s)
	if n.resourceGroup == "" || n.serviceName == "" {
		return nil, fmt.Errorf("failed to get policies, did not find APIM Service ResourceGroup or Service Name in ArmApiAppInfo")
	}

	pager := n.client.NewListByServicePager(n.resourceGroup, n.serviceName, nil)
	return clientsupport.DoListAndMap(pager, simpleMapper, "GetPolicies")
}

func (n *NamedValuePolicyStoreSvc) SetPolicy(aRar rar.ResourceActionRoles) error {
	logger.Info("NamedValuePolicyStoreSvc.SetPolicy", "resourceGroup", n.resourceGroup, "service", n.serviceName)

	nvName := rarToNvName(aRar)
	nvValue := rarToNvValue(aRar)
	updater := n.beginUpdateFunc(context.Background(), n.resourceGroup, n.serviceName, nvName, nvValue)
	cp := clientsupport.NewArmLroPoller(updater)
	_, err := cp.PollForResult(context.Background(), time.Second*5)
	return err
}

func (n *NamedValuePolicyStoreSvc) beginUpdateFunc(ctx context.Context, resourceGroup string, service string, nvName string, nvVal string) clientsupport.GetPollerFunc[azarmapim.NamedValueClientUpdateResponse] {
	updater := func() (*runtime.Poller[azarmapim.NamedValueClientUpdateResponse], error) {
		updateParams := azarmapim.NamedValueUpdateParameters{
			Properties: &azarmapim.NamedValueUpdateParameterProperties{
				Value: &nvVal,
			},
		}
		return n.client.BeginUpdate(ctx, resourceGroup, service, nvName, "*", updateParams, nil)
	}

	return updater
}

func simpleMapper(page azarmapim.NamedValueClientListByServiceResponse) []rar.ResourceActionRoles {
	resRoles := make([]rar.ResourceActionRoles, 0)
	for _, nva := range page.Value {
		nvName := *nva.Name
		nvVal := *nva.Properties.Value
		aRar, err := NvToRar(nvName, nvVal)
		if err != nil {
			logger.Error("NamedValuePolicyStoreSvc.simpleMapper", "msg", "failed to convert namedvalue to rar", "nvName", nvName, "nvVal", nvVal)
			return nil
		}
		resRoles = append(resRoles, aRar)
	}

	return resRoles
}

func NvToRar(nvName, nvVal string) (rar.ResourceActionRoles, error) {
	roles, err := NvValueToArray(nvVal)
	if err != nil {
		logger.Error("NamedValuePolicyStoreSvc.simpleMapper", "msg", "failed to map NamedValue from azure to roles array", "nv.Name", nvName, "nv.Value", nvVal, "err", err)
		return rar.ResourceActionRoles{}, err
	}

	nvKeyParts := strings.Split(nvName, nvKeySep)
	if len(nvKeyParts) < 3 {
		logger.Error("NamedValuePolicyStoreSvc.simpleMapper", "msg", "failed to map NamedValue key from azure to rar. Key must have min 3 parts separated by '-'", "nv.Name", nvName)
		return rar.ResourceActionRoles{}, nil
	}

	httpMethod := getHttpMethod(nvKeyParts[1], providerKeyActionPrefix)
	if httpMethod == "" {
		logger.Error("NamedValuePolicyStoreSvc.simpleMapper", "msg", "failed to extract http method from azure NamedValue", "nv.Name", nvName)
		return rar.ResourceActionRoles{}, nil
	}

	resource := strings.Join(nvKeyParts[2:], "/")
	if resource == "" {
		logger.Error("NamedValuePolicyStoreSvc.simpleMapper", "msg", "failed to extract resource from azure NamedValue", "nv.Name", nvName)
		return rar.ResourceActionRoles{}, nil
	}

	aRar, err := rar.NewResourceActionRoles("/"+resource, []string{httpMethod}, roles)
	if err != nil {
		logger.Error("NamedValuePolicyStoreSvc.simpleMapper", "failed to map NamedValue key from azure to rar", err)
		return rar.ResourceActionRoles{}, err
	}

	return aRar, nil
}

func NvValueToArray(nvValue string) ([]string, error) {
	arr := make([]string, 0)
	err := json.Unmarshal([]byte(nvValue), &arr)
	return arr, err
}

// getHttpMethod - converts an action ("httpget") or actionUri("http:GET") e.g.  to the
// corresponding http method i.e. GET
func getHttpMethod(action, actionPrefix string) string {
	for _, httpMethod := range []string{http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace} {

		prefixedHttpMethod := actionPrefix + httpMethod
		if strings.ToLower(prefixedHttpMethod) == strings.ToLower(action) {
			return httpMethod
		}
	}
	return ""
}

func rarToNvName(aRar rar.ResourceActionRoles) string {
	logger.Info("NamedValuePolicyStoreSvc.rarToNvName", "resource", aRar.Resource(), "action", aRar.Actions(), "members", aRar.Members())
	resource := aRar.Resource()

	if strings.TrimSpace(resource) == "" || len(aRar.Actions()) == 0 || strings.TrimSpace(aRar.Actions()[0]) == "" {
		logger.Warn("makeRarKey empty resource or action", "resource", resource, "action", aRar.Actions())
		return ""
	}

	action := aRar.Actions()[0]
	resNoPrefix, _ := strings.CutPrefix(resource, "/")
	httpMethod := getHttpMethod(action, "")
	if httpMethod == "" {
		logger.Warn("MakeRarKey could not resolve httpMethod", "action", action, "resource", resource)
		return ""
	}

	parts := []string{
		rarNVPrefix,
		providerKeyActionPrefix + strings.ToLower(httpMethod),
		strings.ReplaceAll(resNoPrefix, "/", nvKeySep),
	}

	nvName := strings.Join(parts, nvKeySep)
	return nvName
}

// rarToNvValue
// returns a json string representing the roles array
func rarToNvValue(aRar rar.ResourceActionRoles) string {
	nvVal, _ := json.Marshal(aRar.Members())
	return string(nvVal)
}
