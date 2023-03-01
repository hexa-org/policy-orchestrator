package openpolicyagent_test

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"

	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	assert "github.com/stretchr/testify/require"
)

func TestNewGCPBundleClient(t *testing.T) {
	key := []byte(`{"type": "service_account" }`)

	client, err := openpolicyagent.NewGCPBundleClient(
		"bucket",
		"bundle.tar.gz",
		key,
	)

	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestNewGCPBundleClientError_MissingRequired(t *testing.T) {
	key := []byte(`{"type": "service_account"}`)
	client, err := openpolicyagent.NewGCPBundleClient(
		"",
		"bundle.tar.gz",
		key,
	)

	assert.Nil(t, client)
	assert.EqualError(t, err, "required config: bundle_url, bucket_name, object_name")

	client, err = openpolicyagent.NewGCPBundleClient(
		"bucket",
		"",
		key,
	)

	assert.Nil(t, client)
	assert.EqualError(t, err, "required config: bundle_url, bucket_name, object_name")

	key = []byte(`{}`)
	client, err = openpolicyagent.NewGCPBundleClient(
		"bucket",
		"bundle.tar.gz",
		key,
	)

	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestGCPBundleClient_GetDataFromBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	bucketName := "testBucket"
	objectName := "testbundle.tar.gz"
	key := []byte(`{"type": "service_account"}`)

	mockClient := testsupport.NewMockHTTPClient()
	url := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s?alt=media", bucketName, objectName)
	mockClient.ResponseBody[url] = buffer.Bytes()

	client, err := openpolicyagent.NewGCPBundleClient(
		bucketName,
		objectName,
		key,
		openpolicyagent.WithHTTPClient(mockClient),
	)

	assert.NoError(t, err)
	assert.NotNil(t, client)

	dir := t.TempDir()
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(dir, fmt.Sprintf("test-bundles/.bundle-%d", rand.Uint64()))
	data, err := client.GetDataFromBundle(path)

	assert.NoError(t, err)
	assert.Equal(t, expectedBundleData, string(data))
}

func TestGCPBundleClient_GetDataFromBundleError(t *testing.T) {
	bucketName := "testBucket"
	objectName := "testBundle.tar.gz"
	key := []byte(`{"type": "service_account"}`)

	mockClient := testsupport.NewMockHTTPClient()
	mockClient.Err = errors.New("err getting bundle object")

	client, err := openpolicyagent.NewGCPBundleClient(
		bucketName,
		objectName,
		key,
		openpolicyagent.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	data, err := client.GetDataFromBundle("/some/path")

	assert.Empty(t, data)
	assert.EqualError(t, err, "err getting bundle object")
}

func TestGCPBundleClient_PostBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	bucketName := "test-Bucket"
	objectName := "testBundle.tar.gz"
	key := []byte(`{"type": "service_account"}`)

	mockClient := testsupport.NewMockHTTPClient()
	getObjectMetadataURL := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s", bucketName, objectName)
	mockClient.ResponseBody[getObjectMetadataURL] = []byte(`{"generation": "1234567890"}`)

	postObjectURL, err := url.Parse(fmt.Sprintf("https://storage.googleapis.com/upload/storage/v1/b/%s/o", bucketName))
	assert.NoError(t, err)
	query := postObjectURL.Query()
	query.Add("name", objectName)
	query.Add("uploadType", "media")
	query.Add("ifGenerationMatch", "1234567890")
	postObjectURL.RawQuery = query.Encode()
	mockClient.ResponseBody[postObjectURL.String()] = []byte(`{"generation"": "4567890123"}`)
	mockClient.StatusCode = http.StatusCreated

	client, err := openpolicyagent.NewGCPBundleClient(
		bucketName,
		objectName,
		key,
		openpolicyagent.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	code, err := client.PostBundle(buffer.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, code)
}

func TestGCPBundleClient_PostBundleError(t *testing.T) {
	bucketName := "test-Bucket"
	objectName := "testBundle.tar.gz"
	key := []byte(`{"type": "service_account"}`)

	mockClient := testsupport.NewMockHTTPClient()
	getObjectMetadataURL := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s", bucketName, objectName)

	mockClient.ResponseBody[getObjectMetadataURL] = []byte(`{"generation": "1234567890"}`)

	postObjectURL, err := url.Parse(fmt.Sprintf("https://storage.googleapis.com/upload/storage/v1/b/%s/o", bucketName))
	assert.NoError(t, err)
	query := postObjectURL.Query()
	query.Add("name", objectName)
	query.Add("uploadType", "media")
	query.Add("ifGenerationMatch", "1234567890")
	postObjectURL.RawQuery = query.Encode()
	errorResponse := `
{
  "error": {
    "code": 400,
    "message": "Invalid long value: '%!d(string=1667527988589221)'.",
    "errors": [
      {
        "message": "Invalid long value: '%!d(string=1667527988589221)'.",
        "location": "ifGenerationMatch"
      }
    ]
  }
}
`
	mockClient.ResponseBody[postObjectURL.String()] = []byte(errorResponse)
	mockClient.StatusCode = http.StatusBadRequest

	client, err := openpolicyagent.NewGCPBundleClient(
		bucketName,
		objectName,
		key,
		openpolicyagent.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	code, err := client.PostBundle([]byte(""))

	assert.Error(t, err)
	assert.EqualError(t, err, "error response from GCS: ifGenerationMatch: Invalid long value: '%!d(string=1667527988589221)'.")
	assert.Equal(t, http.StatusInternalServerError, code)
}
