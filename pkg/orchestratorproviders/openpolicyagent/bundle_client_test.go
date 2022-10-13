package openpolicyagent_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent/test"
	"github.com/stretchr/testify/assert"
)

func TestBundleClient_GetExpressionFromBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	mockClient := openpolicyagent_test.MockClient{Response: buffer.Bytes()}

	client := openpolicyagent.BundleClient{BundleServerURL: "someURL", HttpClient: &mockClient}

	dir := os.TempDir()
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(dir, fmt.Sprintf("test-bundles/.bundle-%d", rand.Uint64()))
	data, err := client.GetDataFromBundle(path)
	assert.NoError(t, err)
	assert.Equal(t, `{
  "policies": [
    {
      "meta": {"version": "0.5"},
      "actions": [{"action_uri": "http:GET:/"}],
      "subject": {
        "members": [
          "allusers", "allauthenticated"
        ]
      },
      "object": {
        "resource_id": "aResourceId"
      }
    },
    {
      "meta": {"version": "0.5"},
      "actions": [{"action_uri": "http:GET:/sales"}, {"action_uri": "http:GET:/marketing"}],
      "subject": {
        "members": [
          "allauthenticated",
          "sales@hexaindustries.io",
          "marketing@hexaindustries.io"
        ]
      },
      "object": {
        "resource_id": "aResourceId"
      }
    },
    {
      "meta": {"version": "0.5"},
      "actions": [{"action_uri": "http:GET:/accounting"}, {"action_uri": "http:POST:/accounting"}],
      "subject": {
        "members": [
          "accounting@hexaindustries.io"
        ]
      },
      "object": {
        "resource_id": "aResourceId"
      }
    },
    {
      "meta": {"version": "0.5"},
      "actions": [{"action_uri": "http:GET:/humanresources"}],
      "subject": {
        "members": [
          "humanresources@hexaindustries.io"
        ]
      },
      "object": {
        "resource_id": "aResourceId"
      }
    }
  ]
}`, string(data))
	_ = os.RemoveAll(path)
}

func TestBundleClient_GetExpressionFromBundle_withBadRequest(t *testing.T) {
	mockClient := openpolicyagent_test.MockClient{}
	mockClient.Err = errors.New("oops")
	client := openpolicyagent.BundleClient{BundleServerURL: "someURL", HttpClient: &mockClient}
	_, err := client.GetDataFromBundle(os.TempDir())
	assert.Error(t, err)
}

func TestBundleClient_GetExpressionFromBundle_withBadGzip(t *testing.T) {
	var buffer bytes.Buffer
	mockClient := openpolicyagent_test.MockClient{Response: buffer.Bytes()}
	client := openpolicyagent.BundleClient{BundleServerURL: "someURL", HttpClient: &mockClient}
	_, err := client.GetDataFromBundle("")
	assert.Error(t, err)
}

func TestBundleClient_GetExpressionFromBundle_withBadTar(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	mockClient := openpolicyagent_test.MockClient{Response: buffer.Bytes()}
	client := openpolicyagent.BundleClient{BundleServerURL: "someURL", HttpClient: &mockClient}
	_, err := client.GetDataFromBundle("/badPath")
	assert.Error(t, err)
}
