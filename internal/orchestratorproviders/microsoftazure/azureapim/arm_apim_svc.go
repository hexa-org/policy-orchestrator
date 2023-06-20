package azureapim

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim/armclientsupport"
	log "golang.org/x/exp/slog"
)

type ArmApimSvc interface {
	GetApimServiceInfo(serviceUrl string) (armmodel.ApimServiceInfo, error)
	GetResourceRoles(s armmodel.ApimServiceInfo) ([]armmodel.ResourceActionRoles, error)
	//getApimApiInfo(armResource armmodel.ArmResource, serviceUrl string) (*ApimServiceInfo, error)
}

type armApimSvc struct {
	apimApiClient     ArmApimApiClient
	apimServiceClient ApimServiceClient
	namedValuesClient apimnv.NamedValuesClient
}

func NewArmApimSvc(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (ArmApimSvc, error) {
	apimApiClient, _ := newApimApiClient(subscriptionID, credential, options)
	serviceClient, _ := newArmServiceClient(subscriptionID, credential, options)
	namedValuesClient := apimnv.NewNamedValuesClient(subscriptionID, credential, options)

	return &armApimSvc{
		apimApiClient:     apimApiClient,
		apimServiceClient: serviceClient,
		namedValuesClient: namedValuesClient}, nil
}

func (svc *armApimSvc) GetApimServiceInfo(serviceUrl string) (armmodel.ApimServiceInfo, error) {
	if serviceUrl == "" {
		return armmodel.ApimServiceInfo{}, nil
	}

	pager := svc.apimServiceClient.List(nil)
	mapper := apimServiceInfoMapper(serviceUrl)
	//pageMapper := armclientsupport.NewArmListPageMapper(pager, mapper, "GetApimServiceInfo")
	//pages, err := pageMapper.Get()
	//if err != nil || len(pages) == 0 {
	//	return armmodel.ApimServiceInfo{}, err
	//}

	services, err := doListAndMap(pager, mapper, "GetApimServiceInfo")
	if err != nil || len(services) == 0 {
		return armmodel.ApimServiceInfo{}, err
	}
	return services[0], nil
}

// apimServiceInfoMapper - maps azapim.ServiceClientListResponse to armmodel.ApimServiceInfo
// filters by serviceUrl
// returns empty slice if no apim services found, or if no services match the serviceUrl
func apimServiceInfoMapper(serviceUrl string) func(page azapim.ServiceClientListResponse) []armmodel.ApimServiceInfo {
	return func(page azapim.ServiceClientListResponse) []armmodel.ApimServiceInfo {
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

func (svc *armApimSvc) GetResourceRoles(s armmodel.ApimServiceInfo) ([]armmodel.ResourceActionRoles, error) {
	log.Info("GetResourceRoles", "service", s)
	if s.ResourceGroup == "" || s.Name == "" {
		return []armmodel.ResourceActionRoles{}, nil
	}

	pager := svc.namedValuesClient.List(s.ArmResource.ResourceGroup, s.ArmResource.Name, nil)
	mapper := apimResourceRolesMapper()
	return doListAndMap(pager, mapper, "GetResourceRoles")
	//pageMapper := armclientsupport.NewArmListPageMapper(pager, mapper, "GetResourceRoles")
	//resRoles, err := pageMapper.Get()
	//if err != nil || len(resRoles) == 0 {
	//	return []armmodel.ResourceActionRoles{}, err
	//}
	//return resRoles, nil
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

func apimResourceRolesMapper() func(page azapim.NamedValueClientListByServiceResponse) []armmodel.ResourceActionRoles {
	return func(page azapim.NamedValueClientListByServiceResponse) []armmodel.ResourceActionRoles {
		resRoles := make([]armmodel.ResourceActionRoles, 0)
		for _, nv := range page.Value {
			roles, err := nvValueToArray(*nv.Properties.Value)
			if err != nil {
				log.Info("ignoring apim.NamedValue non-array value", "ValueStr", *nv.Properties.Value, "Err", err)
			}
			one := armmodel.NewResourceActionRoles(*nv.Name, roles)
			resRoles = append(resRoles, one)
		}

		return resRoles
	}
}

/*
func (s *armApimSvc) GetTestInfo(serviceUrl string) (*armmodel.ApimServiceInfo, error) {
	pager := s.apimServiceClient.List(nil)
	mapperFunc := func(page azapim.ServiceClientListResponse) []*armmodel.ApimServiceInfo {
		if len(page.Value) == 0 {
			return nil
		}

		s := page.Value[0]
		if *s.Properties.GatewayURL == serviceUrl {
			return []*armmodel.ApimServiceInfo{
				armmodel.NewApimServiceInfo(*s.ID, *s.Type, *s.Name, *s.Name, *s.Properties.GatewayURL),
			}
			//return armmodel.NewApimServiceInfo(*service.ID, *service.Name, *service.Name, *service.Properties.GatewayURL, *service.Type), nil
		}
		return nil
	}

	myPager := armclientsupport.NewArmListPageMapper[azapim.ServiceClientListResponse, *armmodel.ApimServiceInfo](pager, mapperFunc, "GetTestInfo")

	pages, err := myPager.Get()
	if err != nil || len(pages) == 0 {
		return nil, err
	}

	return pages[0], nil
}
*/
