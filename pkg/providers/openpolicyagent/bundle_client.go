package openpolicyagent

import (
	"bytes"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type BundleClient struct {
	HttpClient HTTPClient
}

func (b *BundleClient) GetExpressionFromBundle(bundleUrl string, path string) ([]byte, error) {
	get, getErr := b.HttpClient.Get(bundleUrl)
	if getErr != nil {
		return nil, getErr
	}

	all, readErr := io.ReadAll(get.Body)
	if readErr != nil {
		return nil, readErr
	}

	gz, gzipErr := compressionsupport.UnGzip(bytes.NewReader(all))
	if gzipErr != nil {
		return nil, gzipErr
	}

	tarErr := compressionsupport.UnTarToPath(bytes.NewReader(gz), path)
	if tarErr != nil {
		return nil, tarErr
	}

	join := filepath.Join(path, "bundle/policy.rego")

	return os.ReadFile(join)
}
