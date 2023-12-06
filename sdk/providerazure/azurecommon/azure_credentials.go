package azurecommon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	log "golang.org/x/exp/slog"
)

type AzureKey struct {
	AppId        string `json:"appId"`
	Secret       string `json:"secret"`
	Tenant       string `json:"tenant"`
	Subscription string `json:"subscription"`
}

func ClientSecretCredentials(key []byte, httpClient HTTPClient) (*azidentity.ClientSecretCredential, error) {
	azureKey, err := DecodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure credentials from key. error: %w", err)
	}

	return ClientSecretCredentialsWithAzureKey(azureKey, httpClient)
}

func ClientSecretCredentialsWithAzureKey(azureKey AzureKey, httpClient HTTPClient) (*azidentity.ClientSecretCredential, error) {
	credOpts := credentialOptions(httpClient)
	credentials, err := azidentity.NewClientSecretCredential(azureKey.Tenant, azureKey.AppId, azureKey.Secret, credOpts)

	if err != nil {
		log.Error("error creating azure credential", err)
		return nil, err
	}

	return credentials, nil
}

func DecodeKey(key []byte) (AzureKey, error) {
	var decoded AzureKey
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&decoded)
	if err != nil {
		return AzureKey{}, fmt.Errorf("failed to decode azure key. error: %w", err)
	}
	return decoded, err
}

func credentialOptions(httpClient HTTPClient) *azidentity.ClientSecretCredentialOptions {
	var credOpts *azidentity.ClientSecretCredentialOptions
	if httpClient != nil {
		credOpts = &azidentity.ClientSecretCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Retry:     policy.RetryOptions{MaxRetries: -1},
				Transport: httpClient,
			},
		}
	}
	return credOpts
}
