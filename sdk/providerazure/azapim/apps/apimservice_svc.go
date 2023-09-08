package apps

import (
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/apps/internal/client"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/clientsupport"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/model"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azurecommon"
)

type appInfoSvc struct {
	svcClient client.ApimServiceClient
}

func NewAppInfoSvc(key []byte, httpClient azurecommon.HTTPClient) (idp.AppInfoSvc, error) {
	svcClient, err := client.NewApimServiceClient(key, httpClient)
	if err != nil {
		return nil, err
	}
	return &appInfoSvc{svcClient: svcClient}, nil
}

func (as *appInfoSvc) GetApplications() ([]idp.AppInfo, error) {
	return as.getApplications()
}

func (as *appInfoSvc) GetApplication(key string) (idp.AppInfo, error) {
	//return as.svcClient.Get(context.TODO(), as.)
	return nil, nil
}

func (as *appInfoSvc) getApplications() ([]idp.AppInfo, error) {
	pager := as.svcClient.NewListPager(nil)

	services, err := clientsupport.DoListAndMap(pager, apimServiceInfoMapper, "GetApplications")
	if err != nil || len(services) == 0 {
		return nil, err
	}
	return services, nil
}

// apimServiceInfoMapper - maps azarmapim.ServiceClientListResponse to armmodel.ApimServiceInfo
// filters by serviceUrl
// returns empty slice if no apim services found, or if no services match the serviceUrl
func apimServiceInfoMapper(page azarmapim.ServiceClientListResponse) []idp.AppInfo {
	apps := make([]idp.AppInfo, 0)
	if len(page.Value) == 0 {
		return apps
	}

	for _, s := range page.Value {
		aApp := model.NewArmApiAppInfo(*s.ID, *s.Type, *s.Name, *s.Name, *s.Properties.GatewayURL)

		apps = append(apps, aApp)
	}
	return apps
}
