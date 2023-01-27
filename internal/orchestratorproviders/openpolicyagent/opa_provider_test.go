package openpolicyagent_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/openpolicyagent/test"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverApplications(t *testing.T) {
	type expected struct {
		ProjectID string
		ObjectID  string
	}

	tests := []struct {
		name     string
		key      []byte
		expected struct {
			ProjectID string
			ObjectID  string
		}
	}{
		{
			name: "with project id",
			key: []byte(`
              {
                "bundle_url": "aBigUrl",
                "project_id": "some opa project"
              }`),

			expected: expected{ProjectID: "some opa project", ObjectID: base64.StdEncoding.EncodeToString([]byte("aBigUrl"))},
		},
		{
			name: "without project id",
			key: []byte(`
              {
                "bundle_url": "aBigUrl"
              }`),
			expected: expected{ProjectID: "package authz", ObjectID: base64.StdEncoding.EncodeToString([]byte("aBigUrl"))},
		},
		{
			name: "gcp bundle storage project",
			key: []byte(`
{
  "project_id": "some gcp project",
  "gcp": {
    "bucket_name": "opa-bundles",
    "object_name": "bundle.tar.gz",
    "key": {
      "type": "service_account"
    }
  }
}
`),
			expected: expected{ProjectID: "some gcp project", ObjectID: "opa-bundles"},
		},
		{
			name: "aws bundle storage project",
			key: []byte(`
{
  "project_id": "some aws project",
  "aws": {
    "bucket_name": "opa-bundles",
    "object_name": "bundle.tar.gz",
	"key": {
      "region": "us-west-1"
    }
  }
}
`),
			expected: expected{ProjectID: "some aws project", ObjectID: "opa-bundles"},
		},
		{
			name: "github bundle storage project",
			key: []byte(`
{
  "project_id": "some github project",
  "github": {
    "account": "hexa-org",
    "repo": "opa-bundles",
	"bundlePath": "bundle.tar.gz",
	"key": {
      "accessToken": "some_github_access_token"
    }
  }
}
`),
			expected: expected{ProjectID: "some github project", ObjectID: "opa-bundles"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := openpolicyagent.OpaProvider{}

			applications, err := p.DiscoverApplications(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: tt.key})

			assert.NoError(t, err)
			assert.Equal(t, 1, len(applications))
			assert.Equal(t, tt.expected.ProjectID, applications[0].Name)
			assert.Equal(t, tt.expected.ObjectID, applications[0].ObjectID)
			assert.Equal(t, "Open Policy Agent bundle", applications[0].Description)
			assert.Equal(t, "Hexa OPA", applications[0].Service)
		})
	}
}

func TestDiscoverApplications_Error(t *testing.T) {
	p := openpolicyagent.OpaProvider{}

	applications, err := p.DiscoverApplications(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: []byte("bad key")})

	assert.Empty(t, applications)
	assert.Error(t, err)
}

func TestOpaProvider_EnsureClientIsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		key            []byte
		expectedClient any
	}{
		{
			name: "http bundle client",
			key: []byte(`
{
  "bundle_url": "aBigUrl"
}
`),
			expectedClient: &openpolicyagent.HTTPBundleClient{},
		},
		{
			name: "gcp bundle client",
			key: []byte(`
{
  "bundle_url": "aBigUrl",
  "gcp": {
    "bucket_name": "opa-bundles",
    "object_name": "bundle.tar.gz",
    "key": {
      "type": "service_account"
    }
  }
}
`),
			expectedClient: &openpolicyagent.GCPBundleClient{},
		},
		{
			name: "aws bundle client",
			key: []byte(`
{
  "aws": {
    "bucket_name": "opa-bundles",
    "object_name": "bundle.tar.gz",
	"key": {
      "region": "us-west-1"
    }
  }
}
`),
			expectedClient: &openpolicyagent.AWSBundleClient{},
		},
		{
			name: "github bundle client",
			key: []byte(`
{
  "project_id": "some github project",
  "github": {
    "account": "hexa-org",
    "repo": "opa-bundles",
	"bundlePath": "bundle.tar.gz",
	"key": {
      "accessToken": "some_github_access_token"
    }
  }
}
`),
			expectedClient: &openpolicyagent.GithubBundleClient{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, file, _, _ := runtime.Caller(0)

			resourcesDirectory := filepath.Join(file, "../resources")
			p := openpolicyagent.OpaProvider{ResourcesDirectory: resourcesDirectory}

			client, err := p.ConfigureClient(tt.key)

			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.IsType(t, tt.expectedClient, client)
		})
	}
}

func TestOpaProvider_EnsureClientIsAvailable_Error(t *testing.T) {
	key := []byte("bad key")
	p := openpolicyagent.OpaProvider{}

	_, err := p.ConfigureClient(key)

	assert.Contains(t, err.Error(), "invalid integration key")

	key = []byte(`
{
  "bundle_url": "bundleURL",
  "gcp": {"key": {"bad":"key"}},
  "aws": {"key": {"bad":"key"}},
  "github": {"key": {"bad":"key"}}
}
`)
	p = openpolicyagent.OpaProvider{}
	_, err = p.ConfigureClient(key)

	assert.Contains(t, err.Error(), "unable to create GCS storage client")
}

func TestGetPolicyInfo(t *testing.T) {
	key := []byte(`{"bundle_url": "aBigUrl"}`)
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles/bundle/data.json")
	data, _ := os.ReadFile(join)
	m := &openpolicyagent_test.MockBundleClient{GetResponse: data}

	resourcesDirectory := filepath.Join(file, "../resources")
	p := openpolicyagent.OpaProvider{
		BundleClientOverride: m,
		ResourcesDirectory:   resourcesDirectory,
	}

	policies, _ := p.GetPolicyInfo(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key}, orchestrator.ApplicationInfo{})

	assert.Equal(t, 4, len(policies))
}

func TestGetPolicyInfo_withBadKey(t *testing.T) {
	p := openpolicyagent.OpaProvider{
		BundleClientOverride: &openpolicyagent_test.MockBundleClient{},
	}

	_, err := p.GetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: []byte("bad key")},
		orchestrator.ApplicationInfo{},
	)

	assert.Contains(t, err.Error(), "invalid client")
}

func TestGetPolicyInfo_withBadRequest(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	mockClient := &openpolicyagent_test.MockBundleClient{
		GetErr: errors.New("oops"),
	}
	p := openpolicyagent.OpaProvider{BundleClientOverride: mockClient}

	_, err := p.GetPolicyInfo(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key}, orchestrator.ApplicationInfo{})

	assert.Error(t, err)
}

func TestSetPolicyInfo(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	mockClient := &openpolicyagent_test.MockBundleClient{PostStatusCode: http.StatusCreated}
	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: mockClient, ResourcesDirectory: filepath.Join(file, "../resources")}

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

	gzip, _ := compressionsupport.UnGzip(bytes.NewReader(mockClient.ArgPostBundle))
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(file, fmt.Sprintf("../resources/bundles/.bundle-%d", rand.Uint64()))
	defer os.RemoveAll(path)
	_ = compressionsupport.UnTarToPath(bytes.NewReader(gzip), path)
	readFile, _ := os.ReadFile(path + "/bundle/data.json")
	assert.JSONEq(t, `{"policies":[{"meta":{"version":"0.5"},"actions":[{"action_uri":"http:GET"}],"subject":{"members":["allusers"]},"object":{"resource_id":"anotherResourceId"}}]}`, string(readFile))
}

func TestSetPolicyInfo_withInvalidArguments(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	client := &openpolicyagent_test.MockBundleClient{}
	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}

	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key},
		orchestrator.ApplicationInfo{},
		[]policysupport.PolicyInfo{},
	)

	assert.Equal(t, 500, status)
	assert.Contains(t, err.Error(), "invalid app info")

	status, err = p.SetPolicyInfo(
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
	assert.Contains(t, err.Error(), "invalid policy info")

	key = []byte(`{
  "bundle_url": "aBigUrl",
  "gcp": {"key": {}}
}`)
	status, err = p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectID"},
		[]policysupport.PolicyInfo{},
	)

	assert.Equal(t, 500, status)
	assert.Contains(t, err.Error(), "invalid client")
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
	client := &openpolicyagent.HTTPBundleClient{}
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
