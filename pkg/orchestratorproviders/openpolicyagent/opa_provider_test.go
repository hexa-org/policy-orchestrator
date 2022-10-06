package openpolicyagent_test

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent"
	openpolicyagent_test "github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
)

func TestDiscoverApplications(t *testing.T) {
	tests := []struct {
		name              string
		key               []byte
		expectedProjectID string
	}{
		{
			name: "with project id",
			key: []byte(`
              {
                "bundle_url": "aBigUrl",
                "project_id": "some opa project"
              }`),
			expectedProjectID: "some opa project",
		},
		{
			name: "without project id",
			key: []byte(`
              {
                "bundle_url": "aBigUrl",
              }`),
			expectedProjectID: "package authz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := openpolicyagent.OpaProvider{}
			applications, _ := p.DiscoverApplications(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: tt.key})
			assert.Equal(t, 1, len(applications))
			assert.Equal(t, tt.expectedProjectID, applications[0].Name)
			assert.Equal(t, "Open Policy Agent bundle", applications[0].Description)
			assert.Equal(t, "bundle server", applications[0].Service)
		})
	}
}

func TestGetPolicyInfo(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(join)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	mockClient := openpolicyagent_test.MockClient{Response: buffer.Bytes()}
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}

	resourcesDirectory := filepath.Join(file, "../resources")
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: resourcesDirectory}

	policies, _ := p.GetPolicyInfo(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key}, orchestrator.ApplicationInfo{})
	assert.Equal(t, 4, len(policies))
}

func TestGetPolicyInfo_withBadKey(t *testing.T) {
	client := openpolicyagent.BundleClient{}
	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}
	_, err := p.GetPolicyInfo(orchestrator.IntegrationInfo{}, orchestrator.ApplicationInfo{})
	assert.Error(t, err)
}

func TestGetPolicyInfo_withBadRequest(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	mockClient := openpolicyagent_test.MockClient{}
	mockClient.Err = errors.New("oops")
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}
	_, err := p.GetPolicyInfo(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key}, orchestrator.ApplicationInfo{})
	assert.Error(t, err)
}

func TestGetPolicyInfo_withBadResourceDir(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	mockClient := openpolicyagent_test.MockClient{}
	mockClient.Err = errors.New("oops")
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}
	_, err := p.GetPolicyInfo(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key}, orchestrator.ApplicationInfo{})
	assert.Error(t, err)
}

func TestSetPolicyInfo(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	mockClient := openpolicyagent_test.MockClient{Status: http.StatusCreated}
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}

	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anotherResourceId"},
		[]policysupport.PolicyInfo{
			{Meta: policysupport.MetaInfo{Version: "0.5"}, Actions: []policysupport.ActionInfo{{"http:GET"}}, Subject: policysupport.SubjectInfo{Members: []string{"allusers"}}, Object: policysupport.ObjectInfo{
				ResourceID: "aResourceId",
			}},
		},
	)
	assert.Equal(t, http.StatusCreated, status)
	assert.NoError(t, err)

	gzip, _ := compressionsupport.UnGzip(bytes.NewReader(mockClient.Request))
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(file, fmt.Sprintf("../resources/bundles/.bundle-%d", rand.Uint64()))
	_ = compressionsupport.UnTarToPath(bytes.NewReader(gzip), path)
	readFile, _ := ioutil.ReadFile(path + "/bundle/data.json")
	assert.Equal(t, `{"policies":[{"meta":{"version":"0.5"},"actions":[{"action_uri":"http:GET"}],"subject":{"members":["allusers"]},"object":{"resource_id":"anotherResourceId"}}]}`, string(readFile))
	_ = os.RemoveAll(path)
}

func TestSetPolicyInfo_withInvalidArguments(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	mockClient := openpolicyagent_test.MockClient{Status: -1}
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}

	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}
	status, _ := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key},
		orchestrator.ApplicationInfo{},
		[]policysupport.PolicyInfo{},
	)
	assert.Equal(t, 500, status)

	status, _ = p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "aResourceId"},
		[]policysupport.PolicyInfo{
			{
				Actions: []policysupport.ActionInfo{{"http:GET"}}, Subject: policysupport.SubjectInfo{Members: []string{"allusers"}}, Object: policysupport.ObjectInfo{
					ResourceID: "aResourceId",
				}},
		},
	)
	assert.Equal(t, 500, status)
}

func TestSetPolicyInfo_WithHTTPSBundleServer(t *testing.T) {
	mockCalled := false
	bundleServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bundles", r.URL.Path)
		mockCalled = true
		rw.WriteHeader(http.StatusCreated)
	}))
	caCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: bundleServer.Certificate().Raw,
	})

	integration := struct {
		BundleURL string `json:"bundle_url"`
		CACert    string `json:"ca_cert"`
	}{
		BundleURL: bundleServer.URL,
		CACert:    string(caCert),
	}
	key, err := json.Marshal(integration)
	assert.NoError(t, err)

	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{ResourcesDirectory: filepath.Join(file, "../resources")}
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "aResourceId"},
		[]policysupport.PolicyInfo{
			{Meta: policysupport.MetaInfo{Version: "0.5"}, Actions: []policysupport.ActionInfo{{"http:GET"}}, Subject: policysupport.SubjectInfo{Members: []string{"allusers"}}, Object: policysupport.ObjectInfo{
				ResourceID: "aResourceId",
			}},
		},
	)
	assert.Equal(t, http.StatusCreated, status)
	assert.NoError(t, err)
	assert.True(t, mockCalled)
}

func TestMakeDefaultBundle(t *testing.T) {
	client := openpolicyagent.BundleClient{}
	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}

	data := []byte(`{
  "policies": [
    {
      "version": "0.5",
      "action_uri": "http:GET",
      "subject": {
        "members": [
          "allusers",
          "allauthenticated"
        ]
      },
      "object": {
        "resource_id": "aResourceId"
      }
    }
  ]
}`)
	bundle, _ := p.MakeDefaultBundle(data)

	gzip, _ := compressionsupport.UnGzip(bytes.NewReader(bundle.Bytes()))
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(os.TempDir(), fmt.Sprintf("/test-bundle-%d", rand.Uint64()))
	_ = compressionsupport.UnTarToPath(bytes.NewReader(gzip), path)

	created, _ := ioutil.ReadFile(filepath.Join(path, "/bundle/policy.rego"))
	assert.Contains(t, string(created), "package authz")

	mcreated, _ := ioutil.ReadFile(filepath.Join(path, "/bundle/.manifest"))
	assert.Contains(t, string(mcreated), "{\"revision\":\"\",\"roots\":[\"\"]}")

	dcreated, _ := ioutil.ReadFile(filepath.Join(path, "/bundle/data.json"))
	assert.Equal(t, `{
  "policies": [
    {
      "version": "0.5",
      "action_uri": "http:GET",
      "subject": {
        "members": [
          "allusers",
          "allauthenticated"
        ]
      },
      "object": {
        "resource_id": "aResourceId"
      }
    }
  ]
}`, string(dcreated))
	_ = os.RemoveAll(path)
}
