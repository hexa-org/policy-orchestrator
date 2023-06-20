package azureresource

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azurecommon"
	log "golang.org/x/exp/slog"
)

type ArmResourceSvc interface {
	GetApiManagementResources() ([]armmodel.ArmResource, error)
}

type armResourceSvc struct {
	client ArmResourcesClient
}

func NewArmResourceSvc(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (ArmResourceSvc, error) {
	client, err := newArmResourcesClient(subscriptionID, credential, options)
	if err != nil {
		return nil, err
	}
	return &armResourceSvc{client: client}, nil
}

func (s *armResourceSvc) GetApiManagementResources() ([]armmodel.ArmResource, error) {
	filter := "resourceType eq 'Microsoft.ApiManagement/service'"
	opts := armresources.ClientListOptions{Filter: &filter}

	// *runtime.Pager[ClientListResponse]
	pager := s.client.NewListPager(&opts)
	resources := make([]armmodel.ArmResource, 0)

	for pager.More() {

		page, err := pager.NextPage(context.Background())
		if err != nil {
			parsedError := azurecommon.ParseResponseError(err)
			log.Error("error calling GetApiManagementResources", "Error", parsedError)
			return nil, parsedError
		}
		//page.ResourceListResult
		for _, genericRes := range page.ResourceListResult.Value {

			log.Info("GetApiManagementResources", "genericRes", *genericRes)
			/*resBytes, err := genericRes.MarshalJSON()
			if err != nil {
				log.Error("Generic resource", "Error=", err)
			}
			log.Info("GetApiManagementResources", "Generic resource=", string(resBytes))*/

			res, err := armmodel.NewArmResource(*genericRes.ID, *genericRes.Type, *genericRes.Name, *genericRes.Name)
			if err != nil {
				return nil, err
			}

			resources = append(resources, res)
		}
	}

	return resources, nil
}
