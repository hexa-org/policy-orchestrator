package azurecommon

import (
	"io"
	"net/http"
)

// HTTPClient - copied from
// demo/internal/orchestratorproviders/microsoftazure/azurecommon/azure_http_client.go
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}
