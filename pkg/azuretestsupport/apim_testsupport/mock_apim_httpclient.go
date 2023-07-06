package apim_testsupport

import (
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
)

type AzureApimHttpClient struct {
	HttpClient *testsupport.MockHTTPClient
}

func MockApimHttpClient() *AzureApimHttpClient {
	return &AzureApimHttpClient{HttpClient: armtestsupport.FakeTokenCredentialHttpClient(armtestsupport.Issuer)}
}
