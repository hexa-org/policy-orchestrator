package armtestsupport

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"net/http"
)

const ApimResourceGroupName = "test_apim_resource_group_name"
const ApimServiceName = "test_apim_service_name"
const ApimServiceGatewayUrl = "https://my_apim_service_name.azure-api.net"
const ApimAppId = "test_apim_app_id"
const Issuer = "https://staratatest.io"
const ApiVersion = "2021-04-01"

const AzureSubscriptionsBaseUrl = "https://management.azure.com/subscriptions"

// FakeTokenCredentialHttpClient creates a http client that returns a
// fake access token from a fake token endpoint that is determined by a fake well_known endpoint.
func FakeTokenCredentialHttpClient(issuer string) *testsupport.MockHTTPClient {
	httpClient := testsupport.NewMockHTTPClient()

	wellKnownResp := wellKnownConfig(issuer)

	tokenResp := struct {
		AccessToken string
	}{
		AccessToken: "accessToken",
	}

	token, _ := json.Marshal(tokenResp)
	httpClient.AddRequest(http.MethodGet, "https://login.microsoftonline.com/atenant/v2.0/.well-known/openid-configuration", http.StatusOK, wellKnownResp)
	httpClient.AddRequest(http.MethodPost, issuer+"/token", http.StatusOK, token)
	return httpClient
}

func wellKnownConfig(issuer string) []byte {
	wellKnownResp := struct {
		Issuer                string `json:"issuer,omitempty"`
		AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
		TokenEndpoint         string `json:"token_endpoint,omitempty"`
		JWKSEndpoint          string `json:"jwks_uri,omitempty"`
	}{
		Issuer:                issuer,
		AuthorizationEndpoint: fmt.Sprintf("%s/authorize", issuer),
		TokenEndpoint:         fmt.Sprintf("%s/token", issuer),
		JWKSEndpoint:          fmt.Sprintf("%s/jwks", issuer),
	}

	wk, _ := json.Marshal(wellKnownResp)
	return wk
}
