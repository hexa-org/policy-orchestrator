package openpolicyagent_test

import (
	"bytes"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNewAWSBundleClient(t *testing.T) {
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "aRegion"
}
`)

	client, err := openpolicyagent.NewAWSBundleClient(
		"bucket",
		"bundle.tar.gz",
		key,
		openpolicyagent.AWSBundleClientOptions{})

	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestNewAWSBundleClientError_MissingRequired(t *testing.T) {
	tests := []struct {
		name   string
		bucket string
		object string
	}{
		{
			name:   "missing bucket",
			bucket: "",
			object: "bundle.tar.gz",
		},
		{
			name:   "missing object",
			bucket: "bucket",
			object: "",
		},
	}
	key := []byte(`{"region": "us-west-1"}`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := openpolicyagent.NewAWSBundleClient(
				tt.bucket,
				tt.object,
				key,
				openpolicyagent.AWSBundleClientOptions{})

			assert.Nil(t, client)
			assert.EqualError(t, err, "required config: bucket_name, object_name")
		})
	}
}

func TestNewAWSBundleClientError_InvalidCredentialsJson(t *testing.T) {
	key := []byte(`{"badkey"}`)

	client, err := openpolicyagent.NewAWSBundleClient("bucket", "object", key, openpolicyagent.AWSBundleClientOptions{})
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestAWSBundleClient_GetDataFromBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	bucketName := "testBucket"
	objectName := "testbundle.tar.gz"
	region := "us-west-1"
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "us-west-1"
}
`)

	mockClient := testsupport.NewMockHTTPClient()

	expUrl := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s?x-id=GetObject", region, bucketName, objectName)
	mockClient.ResponseBody[expUrl] = buffer.Bytes()

	opt := openpolicyagent.AWSBundleClientOptions{HTTPClient: mockClient, DisableRetry: true}
	client, err := openpolicyagent.NewAWSBundleClient(bucketName, objectName, key, opt)
	assert.NotNil(t, client)
	assert.NoError(t, err)

	dir := t.TempDir()
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(dir, fmt.Sprintf("test-bundles/.bundle-%d", rand.Uint64()))
	data, err := client.GetDataFromBundle(path)
	assert.NoError(t, err)
	assert.Equal(t, expectedBundleData, string(data))
}

func TestAWSBundleClient_GetDataFromBundleError(t *testing.T) {
	bucketName := "testBucket"
	objectName := "testbundle.tar.gz"
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "us-west-1"
}
`)

	client, err := openpolicyagent.NewAWSBundleClient(bucketName, objectName, key, openpolicyagent.AWSBundleClientOptions{DisableRetry: true})
	assert.NoError(t, err)
	assert.NotNil(t, client)

	data, err := client.GetDataFromBundle("/some/path")
	assert.Empty(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to read bundle object from AWS S3 bucket")
}

func TestAWSBundleClient_PostBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	bucketName := "testBucket"
	objectName := "testbundle.tar.gz"
	region := "us-west-1"
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "us-west-1"
}
`)

	mockClient := testsupport.NewMockHTTPClient()
	expUrl := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s?x-id=PutObject", region, bucketName, objectName)
	mockClient.StatusCode = http.StatusCreated
	mockClient.ResponseBody[expUrl] = []byte("{}")

	client, err := openpolicyagent.NewAWSBundleClient(bucketName, objectName, key, openpolicyagent.AWSBundleClientOptions{HTTPClient: mockClient, DisableRetry: true})
	assert.NotNil(t, client)
	assert.NoError(t, err)

	code, err := client.PostBundle(buffer.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, code)
}

func TestAWSBundleClient_PostBundleError(t *testing.T) {
	bucketName := "testBucket"
	objectName := "testbundle.tar.gz"
	key := []byte(`
{
  "accessKeyID": "anAccessKeyID",
  "secretAccessKey": "aSecretAccessKey",
  "region": "us-west-1"
}
`)

	client, err := openpolicyagent.NewAWSBundleClient(bucketName, objectName, key, openpolicyagent.AWSBundleClientOptions{DisableRetry: true})
	assert.NoError(t, err)
	assert.NotNil(t, client)

	data, err := client.PostBundle([]byte("somedata"))
	assert.Equal(t, data, http.StatusInternalServerError)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to write bundle object to AWS S3 bucket")
}
