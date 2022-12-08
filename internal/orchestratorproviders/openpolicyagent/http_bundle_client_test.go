package openpolicyagent_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/openpolicyagent"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

const expectedBundleData = `{
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
}`

func TestHTTPBundleClient_GetExpressionFromBundle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(t.TempDir(), fmt.Sprintf("test-bundles/.bundle-%d", rand.Uint64()))

	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["someURL"] = buffer.Bytes()
	client := openpolicyagent.HTTPBundleClient{BundleServerURL: "someURL", HttpClient: m}

	data, err := client.GetDataFromBundle(path)

	assert.NoError(t, err)
	assert.Equal(t, expectedBundleData, string(data))
}

func TestHTTPBundleClient_GetExpressionFromBundle_withBadTar(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["someURL"] = buffer.Bytes()
	client := openpolicyagent.HTTPBundleClient{BundleServerURL: "someURL", HttpClient: m}

	_, err := client.GetDataFromBundle("/badPath")

	assert.Contains(t, err.Error(), "unable to untar to path")
}

func TestHTTPBundleClient_GetExpressionFromBundle_withBadRequest(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.Err = errors.New("oops")
	client := openpolicyagent.HTTPBundleClient{BundleServerURL: "someURL", HttpClient: m}

	_, err := client.GetDataFromBundle(os.TempDir())

	assert.EqualError(t, err, "oops")
}

func TestHTTPBundleClient_GetExpressionFromBundle_withBadGzip(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["someURL"] = []byte("bad gzip")
	client := openpolicyagent.HTTPBundleClient{BundleServerURL: "someURL", HttpClient: m}

	_, err := client.GetDataFromBundle("")

	assert.Contains(t, err.Error(), "unable to ungzip")
}
