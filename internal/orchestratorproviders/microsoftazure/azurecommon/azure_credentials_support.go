package azurecommon

import (
	"bytes"
	"encoding/json"
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

func ClientSecretCredentials(azureKey AzureKey, httpClient HTTPClient) (*azidentity.ClientSecretCredential, error) {
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

func DecodeKey(key []byte) (AzureKey, error) {
	var decoded AzureKey
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&decoded)
	return decoded, err
}
