package openpolicyagent

import (
	"bytes"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type HTTPBundleClient struct {
	BundleServerURL string
	HttpClient      HTTPClient
}

func (b *HTTPBundleClient) GetDataFromBundle(path string) ([]byte, error) {
	get, getErr := b.HttpClient.Get(b.BundleServerURL)
	if getErr != nil {
		return nil, getErr
	}
	defer get.Body.Close()

	gz, gzipErr := compressionsupport.UnGzip(get.Body)
	if gzipErr != nil {
		return nil, fmt.Errorf("unable to ungzip: %w", gzipErr)
	}

	tarErr := compressionsupport.UnTarToPath(bytes.NewReader(gz), path)
	if tarErr != nil {
		return nil, fmt.Errorf("unable to untar to path: %w", tarErr)
	}

	return os.ReadFile(filepath.Join(path, "/bundle/data.json"))
}

// todo - ignoring errors for the moment while spiking

func (b *HTTPBundleClient) PostBundle(bundle []byte) (int, error) {
	// todo - Log out the errors at minimum.
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	formFile, _ := writer.CreateFormFile("bundle", "bundle.tar.gz")
	_, _ = formFile.Write(bundle)
	_ = writer.Close()
	parse, _ := url.Parse(b.BundleServerURL)
	contentType := writer.FormDataContentType()
	resp, err := b.HttpClient.Post(fmt.Sprintf("%s://%s/bundles", parse.Scheme, parse.Host), contentType, buf)
	return resp.StatusCode, err
}
