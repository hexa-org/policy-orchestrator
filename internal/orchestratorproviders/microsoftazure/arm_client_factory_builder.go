package microsoftazure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armclientsupport"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azresource"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azurecommon"
	log "golang.org/x/exp/slog"
)

type SvcFactory interface {
	NewArmResourceSvc() (azresource.ArmResourceSvc, error)
	NewApimSvc() (azapim.ArmApimSvc, error)
}

type svcFactory struct {
	azureKey    azurecommon.AzureKey
	credentials *azidentity.ClientSecretCredential
	clientOpts  *arm.ClientOptions
}

func NewSvcFactory(key []byte, httpClient azurecommon.HTTPClient) (SvcFactory, error) {
	azureKey, err := azurecommon.DecodeKey(key)
	if err != nil {
		log.Error("error decoding azure key", err)
		return nil, err
	}

	credentials, err := azurecommon.ClientSecretCredentials(azureKey, httpClient)
	if err != nil {
		return nil, err
	}

	clientOptions := armclientsupport.NewArmClientOptions(httpClient)
	return &svcFactory{
		azureKey:    azureKey,
		credentials: credentials,
		clientOpts:  clientOptions,
	}, nil
}

func (sf *svcFactory) NewArmResourceSvc() (azresource.ArmResourceSvc, error) {
	return azresource.NewArmResourceSvc(sf.azureKey.Subscription, sf.credentials, sf.clientOpts)
}

func (sf *svcFactory) NewApimSvc() (azapim.ArmApimSvc, error) {
	return azapim.NewArmApimSvc(sf.azureKey.Subscription, sf.credentials, sf.clientOpts)
}
