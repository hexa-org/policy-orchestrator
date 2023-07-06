package azapim

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
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimapi"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimservice"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"time"

	log "golang.org/x/exp/slog"
)

type ArmApimSvc interface {
	GetApimServiceInfo(serviceUrl string) (armmodel.ApimServiceInfo, error)
	GetResourceRoles(s armmodel.ApimServiceInfo) ([]providerscommon.ResourceActionRoles, error)
	UpdateResourceRole(s armmodel.ApimServiceInfo, nv providerscommon.ResourceActionRoles) error
	//getApimApiInfo(armResource armmodel.ArmResource, serviceUrl string) (*ApimServiceInfo, error)
}

type armApimSvc struct {
	apimApiClient     apimapi.ArmApimApiClient
	apimServiceClient apimservice.Client
	namedValuesClient apimnv.NamedValuesClient
}

type ArmApimSvcOption func(s *armApimSvc)

func WithNamedValuesClient(nvClient apimnv.NamedValuesClient) ArmApimSvcOption {
	return func(s *armApimSvc) {
		s.namedValuesClient = nvClient
	}
}

func WithApimServiceClient(apimServiceClient apimservice.Client) ArmApimSvcOption {
	return func(s *armApimSvc) {
		s.apimServiceClient = apimServiceClient
	}
}

func NewArmApimSvc(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions, opts ...ArmApimSvcOption) (ArmApimSvc, error) {
	apimApiClient := apimapi.NewApimApiClient(subscriptionID, credential, options)
	serviceClient := apimservice.NewClient(subscriptionID, credential, options)
	namedValuesClient := apimnv.NewNamedValuesClient(subscriptionID, credential, options)

	svc := &armApimSvc{
		apimApiClient:     apimApiClient,
		apimServiceClient: serviceClient,
		namedValuesClient: namedValuesClient}

	for _, opt := range opts {
		opt(svc)
	}

	return svc, nil
}

func (svc *armApimSvc) GetApimServiceInfo(serviceUrl string) (armmodel.ApimServiceInfo, error) {
	if serviceUrl == "" {
		return armmodel.ApimServiceInfo{}, nil
	}

	pager := svc.apimServiceClient.NewListPager(nil)
	mapper := apimServiceInfoMapper(serviceUrl)

	services, err := doListAndMap(pager, mapper, "GetApimServiceInfo")
	if err != nil || len(services) == 0 {
		return armmodel.ApimServiceInfo{}, err
	}
	return services[0], nil
}

func (svc *armApimSvc) GetResourceRoles(s armmodel.ApimServiceInfo) ([]providerscommon.ResourceActionRoles, error) {
	log.Info("GetResourceRoles", "service", s)
	if s.ResourceGroup == "" || s.Name == "" {
		return []providerscommon.ResourceActionRoles{}, nil
	}

	pager := svc.namedValuesClient.NewListByServicePager(s.ArmResource.ResourceGroup, s.ArmResource.Name, nil)
	mapper := apimResourceRolesMapper()
	return doListAndMap(pager, mapper, "GetResourceRoles")
}

func (svc *armApimSvc) UpdateResourceRole(s armmodel.ApimServiceInfo, nv providerscommon.ResourceActionRoles) error {
	log.Info("UpdateResourceRole", "service", s)
	if s.ResourceGroup == "" || s.Name == "" {
		return errors.New("UpdateResourceRole resourceGroup or service Name is null")
	}

	updater := svc.beginUpdateFunc(context.Background(), s.ResourceGroup, s.Name, nv.Name(), nv.Value())
	cp := armclientsupport.NewArmLroPoller(updater)
	_, err := cp.PollForResult(context.Background(), time.Second*5)
	return err
}

func (svc *armApimSvc) beginUpdateFunc(ctx context.Context, resourceGroup string, service string, nvName string, nvVal string) armclientsupport.GetPollerFunc[azarmapim.NamedValueClientUpdateResponse] {
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

// apimServiceInfoMapper - maps azarmapim.ServiceClientListResponse to armmodel.ApimServiceInfo
// filters by serviceUrl
// returns empty slice if no apim services found, or if no services match the serviceUrl
func apimServiceInfoMapper(serviceUrl string) func(page azarmapim.ServiceClientListResponse) []armmodel.ApimServiceInfo {
	return func(page azarmapim.ServiceClientListResponse) []armmodel.ApimServiceInfo {
		if len(page.Value) == 0 {
			return []armmodel.ApimServiceInfo{}
		}

		s := page.Value[0]
		if *s.Properties.GatewayURL == serviceUrl {
			return []armmodel.ApimServiceInfo{
				armmodel.NewApimServiceInfo(*s.ID, *s.Type, *s.Name, *s.Name, *s.Properties.GatewayURL),
			}
		}
		return []armmodel.ApimServiceInfo{}
	}
}

func apimResourceRolesMapper() func(page azarmapim.NamedValueClientListByServiceResponse) []providerscommon.ResourceActionRoles {
	return func(page azarmapim.NamedValueClientListByServiceResponse) []providerscommon.ResourceActionRoles {
		resRoles := make([]providerscommon.ResourceActionRoles, 0)
		for _, nv := range page.Value {
			roles, err := nvValueToArray(*nv.Properties.Value)
			if err != nil {
				log.Info("ignoring apim.NamedValue non-array value", "ValueStr", *nv.Properties.Value, "Err", err)
			}
			one := providerscommon.NewResourceActionRolesFromProviderValue(*nv.Name, roles)
			resRoles = append(resRoles, one)
		}

		return resRoles
	}
}

func doListAndMap[T any, R any](p *runtime.Pager[T], m func(page T) []R, caller string) ([]R, error) {
	pageMapper := armclientsupport.NewArmListPageMapper(p, m, caller)
	resRoles, err := pageMapper.Get()
	if err != nil || len(resRoles) == 0 {
		return []R{}, err
	}
	return resRoles, nil
}

func nvValueToArray(nvValue string) ([]string, error) {
	arr := make([]string, 0)
	err := json.Unmarshal([]byte(nvValue), &arr)
	return arr, err
}
