package openpolicyagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/compressionsupport"

	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"google.golang.org/api/option"
	ghttp "google.golang.org/api/transport/http"
)

type GCPBundleClient struct {
	bundleServerURL string
	bucketName      string
	objectName      string
	httpClient      HTTPClient
}

type GCPBundleClientOpt func(client *GCPBundleClient)

func WithHTTPClient(c HTTPClient) GCPBundleClientOpt {
	return func(client *GCPBundleClient) {
		client.httpClient = c
	}
}

func NewGCPBundleClient(bucketName, objectName string, key []byte, opts ...GCPBundleClientOpt) (*GCPBundleClient, error) {
	opt := option.WithCredentialsJSON(key)
	gClientOpts := append([]option.ClientOption{
		option.WithScopes("https://www.googleapis.com/auth/devstorage.read_write"),
	}, opt)
	client, _, err := ghttp.NewClient(context.Background(), gClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create GCS storage client: %w", err)
	}

	if len(bucketName) == 0 || len(objectName) == 0 {
		return nil, errors.New("required config: bundle_url, bucket_name, object_name")
	}

	bundleClient := &GCPBundleClient{
		bucketName: bucketName,
		objectName: objectName,
		httpClient: client,
	}

	for _, o := range opts {
		o(bundleClient)
	}

	return bundleClient, nil
}

func (g *GCPBundleClient) GetDataFromBundle(path string) ([]byte, error) {
	url := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s?alt=media", g.bucketName, g.objectName)
	resp, err := g.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gz, gzipErr := compressionsupport.UnGzip(resp.Body)
	if gzipErr != nil {
		return nil, gzipErr
	}

	tarErr := compressionsupport.UnTarToPath(bytes.NewReader(gz), path)
	if tarErr != nil {
		return nil, tarErr
	}
	log.Printf("reading bundle object %q from GCS bucket %q", g.objectName, g.bucketName)
	return os.ReadFile(filepath.Join(path, "/bundle/data.json"))
}

func (g *GCPBundleClient) PostBundle(bundle []byte) (int, error) {
	getObjectMetadataURL := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s", g.bucketName, g.objectName)
	resp, err := g.httpClient.Get(getObjectMetadataURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	metadata := struct {
		Generation string `json:"generation,omitempty"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to decode object metadata: %w", err)
	}

	postObjectURL, err := url.Parse(fmt.Sprintf("https://storage.googleapis.com/upload/storage/v1/b/%s/o", g.bucketName))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to parse POST object URL: %w", err)
	}
	query := postObjectURL.Query()
	query.Add("name", g.objectName)
	query.Add("uploadType", "media")
	query.Add("ifGenerationMatch", metadata.Generation)
	postObjectURL.RawQuery = query.Encode()

	contentType := http.DetectContentType(bundle)
	resp, err = g.httpClient.Post(postObjectURL.String(), contentType, bytes.NewReader(bundle))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to write bundle object to GCS bucket: %w", err)
	}
	defer resp.Body.Close()

	var responseBody GCSAPIErrResp
	_ = json.NewDecoder(resp.Body).Decode(&responseBody)

	if responseBody.Error != nil {
		firstError := responseBody.Error.Errors[0]
		return http.StatusInternalServerError, fmt.Errorf("error response from GCS: %s: %s", firstError.Location, firstError.Message)
	}

	log.Printf("wrote bundle object %q to GCS bucket %q", g.objectName, g.bucketName)
	return http.StatusCreated, nil
}

type GCSAPIErr struct {
	Message  string `json:"message,omitempty"`
	Location string `json:"location,omitempty"`
}

type GCSAPIErrDetail struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Errors  []GCSAPIErr `json:"errors,omitempty"`
}

type GCSAPIErrResp struct {
	Error *GCSAPIErrDetail `json:"error,omitempty"`
}
