package azad

import (
	"context"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azad/internal/client"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azad/internal/model"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
)

type appInfoSvc struct {
	svcClient client.AzureGraphClient
}

func NewAppInfoSvc(svcClient client.AzureGraphClient) *appInfoSvc {
	return &appInfoSvc{svcClient: svcClient}
}

func (as *appInfoSvc) GetApplications() ([]idp.AppInfo, error) {
	return as.getApplications()
}

func (as *appInfoSvc) getApplications() ([]idp.AppInfo, error) {
	// First 100 applications should be good enough
	// Can bump it to 999 without need for paging
	var top int32 = 100
	reqConfig := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
		Headers: nil,
		Options: nil,
		QueryParameters: &applications.ApplicationsRequestBuilderGetQueryParameters{
			Top: &top,
			// Passing select to minimize returned response
			// Can also get "appRoles", but we will do it in the get by id
			Select: []string{"id", "appId", "identifierUris", "displayName"},
		},
	}

	result, err := as.svcClient.Applications().Get(context.TODO(), reqConfig)
	if err != nil {
		return nil, err
	}
	return model.ToAppInfoList(result), nil
}
