package apimnv

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armclientsupport"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	log "golang.org/x/exp/slog"
	"time"
)

type ApimNamedValueSvc interface {
	GetResourceRoles(s armmodel.ApimServiceInfo) ([]providerscommon.ResourceActionRoles, error)
	UpdateResourceRole(s armmodel.ApimServiceInfo, nv providerscommon.ResourceActionRoles) error
}

type apimNamedValueSvc struct {
	namedValuesClient NamedValuesClient
}

type ApimNamedValueSvcOption func(s *apimNamedValueSvc)

func WithNamedValuesClient(nvClient NamedValuesClient) ApimNamedValueSvcOption {
	return func(s *apimNamedValueSvc) {
		s.namedValuesClient = nvClient
	}
}

func NewApimNamedValueSvc(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions, opts ...ApimNamedValueSvcOption) ApimNamedValueSvc {
	client := NewNamedValuesClient(subscriptionID, credential, options)
	svc := &apimNamedValueSvc{namedValuesClient: client}
	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func (svc *apimNamedValueSvc) GetResourceRoles(s armmodel.ApimServiceInfo) ([]providerscommon.ResourceActionRoles, error) {
	log.Info("ApimNamedValueSvc.GetResourceRoles", "service", s)
	if s.ResourceGroup == "" || s.Name == "" {
		return []providerscommon.ResourceActionRoles{}, nil
	}

	pager := svc.namedValuesClient.NewListByServicePager(s.ArmResource.ResourceGroup, s.ArmResource.Name, nil)
	mapper := apimResourceRolesMapper()
	return armclientsupport.DoListAndMap(pager, mapper, "GetResourceRoles")
}

func (svc *apimNamedValueSvc) UpdateResourceRole(s armmodel.ApimServiceInfo, nv providerscommon.ResourceActionRoles) error {
	log.Info("ApimNamedValueSvc.UpdateResourceRole", "service", s)
	if s.ResourceGroup == "" || s.Name == "" {
		return errors.New("UpdateResourceRole resourceGroup or service Name is null")
	}

	updater := svc.beginUpdateFunc(context.Background(), s.ResourceGroup, s.Name, nv.Name(), nv.Value())
	cp := armclientsupport.NewArmLroPoller(updater)
	_, err := cp.PollForResult(context.Background(), time.Second*5)
	return err
}

func (svc *apimNamedValueSvc) beginUpdateFunc(ctx context.Context, resourceGroup string, service string, nvName string, nvVal string) armclientsupport.GetPollerFunc[azarmapim.NamedValueClientUpdateResponse] {
	updater := func() (*runtime.Poller[azarmapim.NamedValueClientUpdateResponse], error) {
		updateParams := azarmapim.NamedValueUpdateParameters{
			Properties: &azarmapim.NamedValueUpdateParameterProperties{
				Value: &nvVal,
			},
		}
		return svc.namedValuesClient.BeginUpdate(ctx, resourceGroup, service, nvName, "*", updateParams, nil)
	}

	return updater
}

func apimResourceRolesMapper() func(page azarmapim.NamedValueClientListByServiceResponse) []providerscommon.ResourceActionRoles {
	return func(page azarmapim.NamedValueClientListByServiceResponse) []providerscommon.ResourceActionRoles {
		resRoles := make([]providerscommon.ResourceActionRoles, 0)
		for _, nv := range page.Value {
			roles, err := nvValueToArray(*nv.Properties.Value)
			if err != nil {
				log.Info("ApimNamedValueSvc.apimResourceRolesMapper ignoring apim.NamedValue non-array value", "ValueStr", *nv.Properties.Value, "Err", err)
			}
			one := providerscommon.NewResourceActionRolesFromProviderValue(*nv.Name, roles)
			resRoles = append(resRoles, one)
		}

		return resRoles
	}
}

func nvValueToArray(nvValue string) ([]string, error) {
	arr := make([]string, 0)
	err := json.Unmarshal([]byte(nvValue), &arr)
	return arr, err
}
