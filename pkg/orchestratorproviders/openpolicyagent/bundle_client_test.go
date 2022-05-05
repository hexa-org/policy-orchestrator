package openpolicyagent_test

import (
	"bytes"
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent/test"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBundleClient_GetExpressionFromBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	mockClient := openpolicyagent_test.MockClient{Response: buffer.Bytes()}

	client := openpolicyagent.BundleClient{HttpClient: &mockClient}

	dir := os.TempDir()
	data, err := client.GetDataFromBundle("someUrl", dir)
	assert.NoError(t, err)
	assert.Equal(t, `{
  "policies": [
    {
      "version": "0.4",
      "action": "GET",
      "subject": {
        "authenticated_users": [
          "allusers"
        ]
      },
      "object": {
        "resources": [
          "/"
        ]
      }
    },
    {
      "version": "0.4",
      "action": "GET",
      "subject": {
        "authenticated_users": [
          "sales@hexaindustries.io",
          "marketing@hexaindustries.io"
        ]
      },
      "object": {
        "resources": [
          "/sales",
          "/marketing"
        ]
      }
    },
    {
      "version": "0.4",
      "action": "GET",
      "subject": {
        "authenticated_users": [
          "accounting@hexaindustries.io"
        ]
      },
      "object": {
        "resources": [
          "/accounting"
        ]
      }
    },
    {
      "version": "0.4",
      "action": "GET",
      "subject": {
        "authenticated_users": [
          "humanresources@hexaindustries.io"
        ]
      },
      "object": {
        "resources": [
          "/humanresources"
        ]
      }
    }
  ]
}`, string(data))
}

func TestBundleClient_GetExpressionFromBundle_withBadRequest(t *testing.T) {
	mockClient := openpolicyagent_test.MockClient{}
	mockClient.Err = errors.New("oops")
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, err := client.GetDataFromBundle("someUrl", os.TempDir())
	assert.Error(t, err)
}

func TestBundleClient_GetExpressionFromBundle_withBadGzip(t *testing.T) {
	var buffer bytes.Buffer
	mockClient := openpolicyagent_test.MockClient{Response: buffer.Bytes()}
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, err := client.GetDataFromBundle("someUrl", "")
	assert.Error(t, err)
}

func TestBundleClient_GetExpressionFromBundle_withBadTar(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	mockClient := openpolicyagent_test.MockClient{Response: buffer.Bytes()}

	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, err := client.GetDataFromBundle("someUrl", "/badPath")
	assert.Error(t, err)
}
