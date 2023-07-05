package apim_testsupport

import (
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport/armtestsupport"
)

type AzureApimHttpClient struct {
	HttpClient *testsupport.MockHTTPClient
}

func MockApimHttpClient() *AzureApimHttpClient {
	return &AzureApimHttpClient{HttpClient: armtestsupport.MockAuthorizedHttpClient(armtestsupport.Issuer)}
}
