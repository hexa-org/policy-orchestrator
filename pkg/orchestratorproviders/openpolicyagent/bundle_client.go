package openpolicyagent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"google.golang.org/api/option"
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

type GCPBundleClient struct {
	bundleServerURL string
	bucketName      string
	objectName      string
	gcpClient       *storage.Client
}

func NewGCPBundleClient(bundleURL, bucketName, objectName string, key []byte) (*GCPBundleClient, error) {
	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(key))
	if err != nil {
		return nil, fmt.Errorf("unable to create gcp storage client: %w", err)
	}
	if len(bundleURL) == 0 || len(bucketName) == 0 || len(objectName) == 0 {
		return nil, errors.New("required config: bundle_url, bucket_name, object_name")
	}

	return &GCPBundleClient{
		bundleServerURL: bundleURL,
		bucketName:      bucketName,
		objectName:      objectName,
		gcpClient:       client,
	}, nil
}

// todo - how to test the following? gcpClient is all struct based and difficult to mock.

func (g *GCPBundleClient) GetDataFromBundle(path string) ([]byte, error) {
	r, err := g.gcpClient.Bucket(g.bucketName).Object(g.objectName).NewReader(context.Background())
	if err != nil {
		return nil, err
	}
	defer r.Close()

	gz, gzipErr := compressionsupport.UnGzip(r)
	if gzipErr != nil {
		return nil, gzipErr
	}

	tarErr := compressionsupport.UnTarToPath(bytes.NewReader(gz), path)
	if tarErr != nil {
		return nil, tarErr
	}
	log.Printf("reading bundle object %q from GCP bucket %q", g.objectName, g.bucketName)
	return os.ReadFile(filepath.Join(path, "/bundle/data.json"))
}

// todo - should this accept an io.Writer?
func (g *GCPBundleClient) PostBundle(bundle []byte) (int, error) {
	objHandle := g.gcpClient.Bucket(g.bucketName).Object(g.objectName)
	ctx := context.Background()

	// If the live object already exists in your bucket, set instead a
	// generation-match precondition using the live object's generation number.
	attrs, err := objHandle.Attrs(ctx)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to get GCP object attribuites: %w", err)
	}
	objHandle.If(storage.Conditions{GenerationMatch: attrs.Generation})

	timedCtx, _ := context.WithTimeout(ctx, time.Minute)
	objWriter := objHandle.NewWriter(timedCtx)

	_, err = io.Copy(objWriter, bytes.NewReader(bundle))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to write bundle: %w", err)
	}
	err = objWriter.Close()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to close writer bundle: %w", err)
	}

	log.Printf("wrote bundle object %q to GCP bucket %q", g.objectName, g.bucketName)
	return http.StatusCreated, nil
}
