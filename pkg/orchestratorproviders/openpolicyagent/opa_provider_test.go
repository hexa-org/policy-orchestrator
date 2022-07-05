package openpolicyagent_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/openpolicyagent/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestDiscoverApplications(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	p := openpolicyagent.OpaProvider{}
	applications, _ := p.DiscoverApplications(orchestrator.IntegrationInfo{Name: "open_policy_agent", Key: key})
	assert.Equal(t, 1, len(applications))
	assert.Equal(t, "package authz", applications[0].Name)
	assert.Equal(t, "Open policy agent bundle", applications[0].Description)
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
		orchestrator.ApplicationInfo{},
		[]policysupport.PolicyInfo{{Meta: policysupport.MetaInfo{Version: "0.5"}, Actions: []policysupport.ActionInfo{{"GET"}}, Subject: policysupport.SubjectInfo{Members: []string{"allusers"}}, Object: policysupport.ObjectInfo{Resources: []string{"/"}}}},
	)
	assert.Equal(t, http.StatusCreated, status)
	assert.NoError(t, err)

	gzip, _ := compressionsupport.UnGzip(bytes.NewReader(mockClient.Request))
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(file, fmt.Sprintf("../resources/bundles/.bundle-%d", rand.Uint64()))
	_ = compressionsupport.UnTarToPath(bytes.NewReader(gzip), path)
	readFile, _ := ioutil.ReadFile(path + "/bundle/data.json")
	assert.Equal(t, `{"policies":[{"meta":{"version":"0.5"},"actions":[{"action":"GET"}],"subject":{"members":["allusers"]},"object":{"resources":["/"]}}]}`, string(readFile))
	_ = os.RemoveAll(path)
}

func TestSetPolicyInfo_withBadResponse(t *testing.T) {
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
		[]policysupport.PolicyInfo{{Meta: policysupport.MetaInfo{Version: "0.5"}, Actions: []policysupport.ActionInfo{{"GET"}}, Subject: policysupport.SubjectInfo{Members: []string{"allusers"}}, Object: policysupport.ObjectInfo{Resources: []string{"/"}}}},
	)
	assert.Equal(t, 0, status)
}

func TestMakeDefaultBundle(t *testing.T) {
	client := openpolicyagent.BundleClient{}
	_, file, _, _ := runtime.Caller(0)
	p := openpolicyagent.OpaProvider{BundleClientOverride: client, ResourcesDirectory: filepath.Join(file, "../resources")}

	data := []byte(`{
  "policies": [
    {
      "version": "0.5",
      "action": "GET",
      "subject": {
        "members": [
          "allusers",
          "allauthenticated"
        ]
      },
      "object": {
        "resources": [
          "/"
        ]
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
      "action": "GET",
      "subject": {
        "members": [
          "allusers",
          "allauthenticated"
        ]
      },
      "object": {
        "resources": [
          "/"
        ]
      }
    }
  ]
}`, string(dcreated))
	_ = os.RemoveAll(path)
}
