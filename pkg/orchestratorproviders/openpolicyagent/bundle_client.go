package openpolicyagent

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type BundleClient struct {
	BundleServerURL string
	HttpClient      HTTPClient
}

func (b *BundleClient) GetDataFromBundle(path string) ([]byte, error) {
	get, getErr := b.HttpClient.Get(b.BundleServerURL)
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
	return os.ReadFile(filepath.Join(path, "/bundle/data.json"))
}

// todo - ignoring errors for the moment while spiking

func (b *BundleClient) PostBundle(bundle []byte) (int, error) {
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
