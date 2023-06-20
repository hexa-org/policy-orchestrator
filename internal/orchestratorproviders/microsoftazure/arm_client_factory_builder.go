package microsoftazure

import (
	"bytes"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azurecommon"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureresource"
	log "golang.org/x/exp/slog"
)

type SvcFactory interface {
	NewArmResourceSvc() (azureresource.ArmResourceSvc, error)
	NewApimSvc() (azureapim.ArmApimSvc, error)
}

type serviceFactory struct {
	azureKey    azurecommon.AzureKey
	credentials *azidentity.ClientSecretCredential
	clientOpts  *arm.ClientOptions
}

func NewApimProviderSvcFactory(key []byte, httpClient HTTPClient) (SvcFactory, error) {
	azureKey, err := decodeKey(key)
	if err != nil {
		log.Error("error decoding azure key", err)
		return nil, err
	}

	credentials, err := clientSecretCredentials(azureKey, httpClient)
	if err != nil {
		return nil, err
	}

	clientOptions := makeClientOpts(httpClient)
	return &serviceFactory{
		azureKey:    azureKey,
		credentials: credentials,
		clientOpts:  clientOptions,
	}, nil
}

func (sf *serviceFactory) NewArmResourceSvc() (azureresource.ArmResourceSvc, error) {
	return azureresource.NewArmResourceSvc(sf.azureKey.Subscription, sf.credentials, sf.clientOpts)
	//client, err := sf.buildArmResourcesClient()
	//if err != nil {
	//	return nil, err
	//}
	//return azureresource.NewArmResourceSvc(client), nil
}

func (sf *serviceFactory) NewApimSvc() (azureapim.ArmApimSvc, error) {
	return azureapim.NewArmApimSvc(sf.azureKey.Subscription, sf.credentials, sf.clientOpts)
	/*apimApiClient, err := sf.buildApimApiClient()
	if err != nil {
		return nil, err
	}
	serviceClient, err := sf.buildArmServiceClient()
	if err != nil {
		return nil, err
	}
	return azureapim.NewArmApimSvc(apimApiClient, serviceClient), nil

	*/
}

/*
func (sf *serviceFactory) buildApimApiClient() (azureapim.ArmApimApiClient, error) {
	factory, err := armapimanagement.NewClientFactory(sf.azureKey.Subscription, sf.credentials, sf.clientOpts)
	if err != nil {
		log.Error("Error from armapimanagement.NewClientFactory. Error=", err)
		return nil, err
	}

	return azureapim.NewArmApimApiClient(factory.NewAPIClient()), nil
}

func (sf *serviceFactory) buildArmServiceClient() (azureapim.ApimServiceClient, error) {
	factory, err := armapimanagement.NewClientFactory(sf.azureKey.Subscription, sf.credentials, sf.clientOpts)
	if err != nil {
		log.Error("Error from armapimanagement.NewClientFactory. Error=", err)
		return nil, err
	}
	return azureapim.NewArmServiceClient(factory.NewServiceClient()), nil
}


func (sf *serviceFactory) buildArmResourcesClient() (azureresource.ArmResourcesClient, error) {
	factory, err := armresources.NewClientFactory(sf.azureKey.Subscription, sf.credentials, sf.clientOpts)
	if err != nil {
		log.Error("Error arm resources.NewClientFactory.", err)
		return nil, err
	}

	return azureresource.NewArmResourcesClient(factory.NewClient()), nil
}
*/

/*
	func BuildFactories(key []byte, httpClient HTTPClient) (*ArmFactories, error) {
		azureKey, err := decodeKey(key)
		if err != nil {
			log.Error("error decoding azure key", err)
			return nil, err
		}

		credentials, err := clientSecretCredentials(azureKey, httpClient)
		if err != nil {
			return nil, err
		}

		clientOptions := makeClientOpts(httpClient)
		apimClientFactory, err := azureapim.NewClientFactory(azureKey, credentials, clientOptions)
		if err != nil {
			return nil, err
		}

		resourceClientFactory, err := azureresource.NewClientFactory(azureKey, credentials, clientOptions)
		if err != nil {
			return nil, err
		}

		return &ArmFactories{
			apimClientFactory:     apimClientFactory,
			resourceClientFactory: resourceClientFactory,
		}, nil
	}

	func (f *ArmFactories) ArmResourceClientFactory() azureresource.ClientFactory {
		return f.resourceClientFactory
	}
*/
func clientSecretCredentials(azureKey azurecommon.AzureKey, httpClient HTTPClient) (*azidentity.ClientSecretCredential, error) {
	var apimCredOpts *azidentity.ClientSecretCredentialOptions
	if httpClient != nil {
		apimCredOpts = &azidentity.ClientSecretCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Retry:     policy.RetryOptions{MaxRetries: -1},
				Transport: httpClient,
			},
		}
	}

	credentials, err := azidentity.NewClientSecretCredential(azureKey.Tenant, azureKey.AppId, azureKey.Secret, apimCredOpts)

	if err != nil {
		log.Error("error create azure credential", err)
		return nil, err
	}

	return credentials, nil
}

func makeClientOpts(httpClient HTTPClient) *arm.ClientOptions {
	var clientOpts *arm.ClientOptions
	if httpClient != nil {
		clientOpts = &arm.ClientOptions{
			ClientOptions: policy.ClientOptions{
				Retry:     policy.RetryOptions{MaxRetries: -1},
				Transport: httpClient,
			},
		}
	}
	return clientOpts
}

func decodeKey(key []byte) (azurecommon.AzureKey, error) {
	var decoded azurecommon.AzureKey
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&decoded)
	return decoded, err
}
