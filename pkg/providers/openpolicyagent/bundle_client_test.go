package openpolicyagent_test

import (
	"bytes"
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

func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	recorder := httptest.ResponseRecorder{}
	recorder.Body = bytes.NewBuffer(m.response)
	return recorder.Result(), nil
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
