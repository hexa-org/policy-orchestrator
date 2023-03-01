package openpolicyagent_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

const (
	ghTestAccount = "hexa-org"
	ghTestRepo    = "opa-bundles"
	ghTestBundle  = "bundle.tar.gz"
)

var ghTestDefaultOpts = openpolicyagent.GithubBundleClientOptions{}

func TestNewGithubBundleClient(t *testing.T) {
	key := []byte(`
{
	"accessToken": "some-github-token"
}
`)

	client, err := openpolicyagent.NewGithubBundleClient(
		ghTestAccount,
		ghTestRepo,
		ghTestBundle,
		key,
		ghTestDefaultOpts)

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewGithubBundleClient_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		account string
		repo    string
		bundle  string
	}{
		{
			name:    "missing account",
			account: "",
			repo:    ghTestRepo,
			bundle:  ghTestBundle,
		},
		{
			name:    "missing repo",
			account: ghTestAccount,
			repo:    "",
			bundle:  ghTestBundle,
		},
		{
			name:    "missing bundle",
			account: ghTestAccount,
			repo:    ghTestRepo,
			bundle:  "",
		},
	}

	key := []byte(`
{
	"accessToken": "some-github-token"
}
`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := openpolicyagent.NewGithubBundleClient(
				tt.account,
				tt.repo,
				tt.bundle,
				key,
				ghTestDefaultOpts)

			assert.Nil(t, client)
			assert.EqualError(t, err, "required config: account, repo, branch, bundle")
		})
	}
}

func TestNewGithubBundleClient_InvalidCredentialsJson(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{
			name: "missing key",
			key:  "",
		},
		{
			name: "invalid json",
			key:  `{"badkey"}`,
		},
		{
			name: "missing token in key",
			key:  `{"name": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := openpolicyagent.NewGithubBundleClient(
				ghTestAccount,
				ghTestRepo,
				ghTestBundle,
				[]byte(tt.key),
				ghTestDefaultOpts)

			assert.Nil(t, client)
			assert.Error(t, err)
		})
	}
}

func TestGithubBundleClient_GetDataFromBundle(t *testing.T) {

	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	base64Content := base64.StdEncoding.EncodeToString(buffer.Bytes())
	ghPublishInfo := openpolicyagent.GithubPublishInfo{Content: base64Content}
	var contentBuffer bytes.Buffer
	_ = json.NewEncoder(&contentBuffer).Encode(ghPublishInfo)

	key := []byte(`
{
	"accessToken": "some-github-token"
}
`)

	mockClient := testsupport.NewMockHTTPClient()
	expUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", ghTestAccount, ghTestRepo, ghTestBundle)
	contentBytes := contentBuffer.Bytes()
	mockClient.AddRequest(http.MethodGet, expUrl, http.StatusOK, contentBytes)

	opt := openpolicyagent.GithubBundleClientOptions{HTTPClient: mockClient}
	client, err := openpolicyagent.NewGithubBundleClient(
		ghTestAccount,
		ghTestRepo,
		ghTestBundle,
		key,
		opt)

	assert.NoError(t, err)
	assert.NotNil(t, client)

	dir := t.TempDir()
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(dir, fmt.Sprintf("test-bundles/.bundle-%d", rand.Uint64()))
	data, err := client.GetDataFromBundle(path)
	assert.NoError(t, err)
	assert.Equal(t, expectedBundleData, string(data))
}

func TestGithubBundleClient_GetDataFromBundle_NotFoundError(t *testing.T) {

	key := []byte(`
{
	"accessToken": "some-github-token"
}
`)

	mockClient := testsupport.NewMockHTTPClient()
	expUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", ghTestAccount, ghTestRepo, ghTestBundle)
	mockClient.AddRequest(http.MethodGet, expUrl, http.StatusNotFound, []byte("{}"))

	opt := openpolicyagent.GithubBundleClientOptions{HTTPClient: mockClient}
	client, err := openpolicyagent.NewGithubBundleClient(
		ghTestAccount,
		ghTestRepo,
		ghTestBundle,
		key,
		opt)

	assert.NoError(t, err)
	assert.NotNil(t, client)

	data, err := client.GetDataFromBundle("/some/path")
	assert.Nil(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to read bundle from Github: NOT FOUND")
}

func TestGithubBundleClient_GetDataFromBundle_BadRequest(t *testing.T) {
	client, err := openpolicyagent.NewGithubBundleClient(
		"%invalid",
		ghTestRepo,
		ghTestBundle,
		[]byte(`{"accessToken": "some-github-token"}`),
		ghTestDefaultOpts)

	data, err := client.GetDataFromBundle("/some/path")
	assert.Nil(t, data)
	assert.Error(t, err)
}

func TestGithubBundleClient_GetDataFromBundle_GithubBadCredentialsError(t *testing.T) {
	key := []byte(`
{
	"accessToken": "some-github-token"
}
`)

	client, err := openpolicyagent.NewGithubBundleClient(
		ghTestAccount,
		ghTestRepo,
		ghTestBundle,
		key,
		ghTestDefaultOpts)

	assert.NoError(t, err)
	assert.NotNil(t, client)

	data, err := client.GetDataFromBundle("/some/path")
	assert.Empty(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to read bundle from Github")
}

func TestGithubBundleClient_PostBundle_Create(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)
	bundleBytes := buffer.Bytes()

	mockClient := testsupport.NewMockHTTPClient()
	expUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", ghTestAccount, ghTestRepo, ghTestBundle)

	mockClient.AddRequest(http.MethodGet, expUrl, http.StatusNotFound, nil)
	mockClient.AddRequest(http.MethodPut, expUrl, http.StatusCreated, []byte("{}"))

	key := []byte(`{"accessToken": "some-token"}`)
	opt := openpolicyagent.GithubBundleClientOptions{HTTPClient: mockClient}
	client, err := openpolicyagent.NewGithubBundleClient(ghTestAccount, ghTestRepo, ghTestBundle, key, opt)
	assert.NotNil(t, client)
	assert.NoError(t, err)

	code, err := client.PostBundle(bundleBytes)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, code)

	expReqBody, _ := json.Marshal(openpolicyagent.GithubPublishInfo{
		Message: "Update opa bundle",
		Content: base64.StdEncoding.EncodeToString(bundleBytes),
	})
	assert.Equal(t, string(expReqBody), string(mockClient.GetRequestBodyByKey("PUT", expUrl)))
}

func TestGithubBundleClient_PostBundle_Update(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)
	bundleBytes := buffer.Bytes()

	mockClient := testsupport.NewMockHTTPClient()
	expUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", ghTestAccount, ghTestRepo, ghTestBundle)

	mockClient.AddRequest(http.MethodGet, expUrl, http.StatusOK, []byte(`{"sha": "somesha" }`))
	mockClient.AddRequest(http.MethodPut, expUrl, http.StatusOK, []byte("{}"))

	key := []byte(`{"accessToken": "some-token"}`)
	opt := openpolicyagent.GithubBundleClientOptions{HTTPClient: mockClient}
	client, err := openpolicyagent.NewGithubBundleClient(ghTestAccount, ghTestRepo, ghTestBundle, key, opt)
	assert.NotNil(t, client)
	assert.NoError(t, err)

	code, err := client.PostBundle(bundleBytes)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, code)

	expReqBody, _ := json.Marshal(openpolicyagent.GithubPublishInfo{
		Message: "Update opa bundle",
		Content: base64.StdEncoding.EncodeToString(bundleBytes),
		Sha:     "somesha",
	})
	assert.Equal(t, string(expReqBody), string(mockClient.GetRequestBodyByKey("PUT", expUrl)))
}

func TestGithubBundleClient_PostBundleError(t *testing.T) {

	mockClient := testsupport.NewMockHTTPClient()
	expUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", ghTestAccount, ghTestRepo, ghTestBundle)

	mockClient.AddRequest(http.MethodGet, expUrl, http.StatusNotFound, []byte(""))
	mockClient.AddRequest(http.MethodPut, expUrl, http.StatusUnauthorized, []byte(`{"message": "Bad Credentials"}`))

	key := []byte(`{"accessToken": "some-token"}`)
	opt := openpolicyagent.GithubBundleClientOptions{HTTPClient: mockClient}
	client, err := openpolicyagent.NewGithubBundleClient(ghTestAccount, ghTestRepo, ghTestBundle, key, opt)
	assert.NotNil(t, client)
	assert.NoError(t, err)

	_, err = client.PostBundle([]byte("somedata"))
	assert.Contains(t, err.Error(), "unable to write bundle to Github:")
}
