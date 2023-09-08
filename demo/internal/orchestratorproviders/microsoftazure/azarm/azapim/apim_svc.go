package azapim

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armclientsupport"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimapi"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimservice"
)

type ArmApimSvc interface {
	GetApimServiceInfo(serviceUrl string) (armmodel.ApimServiceInfo, error)
}

type armApimSvc struct {
	apimApiClient     apimapi.ArmApimApiClient
	apimServiceClient apimservice.Client
}

type ArmApimSvcOption func(s *armApimSvc)

func WithApimServiceClient(apimServiceClient apimservice.Client) ArmApimSvcOption {
	return func(s *armApimSvc) {
		s.apimServiceClient = apimServiceClient
	}
}

func NewArmApimSvc(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions, opts ...ArmApimSvcOption) (ArmApimSvc, error) {
	apimApiClient := apimapi.NewApimApiClient(subscriptionID, credential, options)
	serviceClient := apimservice.NewClient(subscriptionID, credential, options)

	svc := &armApimSvc{
		apimApiClient:     apimApiClient,
		apimServiceClient: serviceClient}

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

	services, err := armclientsupport.DoListAndMap(pager, mapper, "GetApimServiceInfo")
	if err != nil || len(services) == 0 {
		return armmodel.ApimServiceInfo{}, err
	}
	return services[0], nil
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
