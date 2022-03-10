package openpolicyagent_test

import (
	"bytes"
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/openpolicyagent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type MockClient struct {
	mock.Mock
	response []byte
	err      error
}

func (m *MockClient) Get(_ string) (resp *http.Response, err error) {
	recorder := httptest.ResponseRecorder{}
	recorder.Body = bytes.NewBuffer(m.response)
	return recorder.Result(), m.err
}

func TestBundleClient_GetExpressionFromBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	mockClient := MockClient{response: buffer.Bytes()}

	client := openpolicyagent.BundleClient{HttpClient: &mockClient}

	dir := os.TempDir()
	rego, err := client.GetExpressionFromBundle("someUrl", dir)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(rego), "package authz"))
}

func TestBundleClient_GetExpressionFromBundle_withBadRequest(t *testing.T) {
	mockClient := MockClient{}
	mockClient.err = errors.New("oops")
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, err := client.GetExpressionFromBundle("someUrl", os.TempDir())
	assert.Error(t, err)
}

func TestBundleClient_GetExpressionFromBundle_withBadGzip(t *testing.T) {
	var buffer bytes.Buffer
	mockClient := MockClient{response: buffer.Bytes()}
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, err := client.GetExpressionFromBundle("someUrl", "")
	assert.Error(t, err)
}

func TestBundleClient_GetExpressionFromBundle_withBadTar(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	mockClient := MockClient{response: buffer.Bytes()}

	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, err := client.GetExpressionFromBundle("someUrl", "/badPath")
	assert.Error(t, err)
}
