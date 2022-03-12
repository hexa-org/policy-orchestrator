package openpolicyagent_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/openpolicyagent"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/openpolicyagent/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDiscoverApplications(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	p := openpolicyagent.OpaProvider{}
	applications, _ := p.DiscoverApplications(provider.IntegrationInfo{Name: "open_policy_agent", Key: key})
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
	service := openpolicyagent.OpaService{ResourcesDirectory: resourcesDirectory}
	p := openpolicyagent.OpaProvider{Client: client, Service: service}

	policies, _ := p.GetPolicyInfo(provider.IntegrationInfo{Name: "open_policy_agent", Key: key}, provider.ApplicationInfo{})
	assert.Equal(t, 4, len(policies))
}

func TestGetPolicyInfo_withBadKey(t *testing.T) {
	client := openpolicyagent.BundleClient{}
	service := openpolicyagent.OpaService{}
	p := openpolicyagent.OpaProvider{Client: client, Service: service}
	_, err := p.GetPolicyInfo(provider.IntegrationInfo{}, provider.ApplicationInfo{})
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
	_, file, _, _ := runtime.Caller(0)
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	service := openpolicyagent.OpaService{ResourcesDirectory: filepath.Join(file, "../resources")}
	p := openpolicyagent.OpaProvider{Client: client, Service: service}
	_, err := p.GetPolicyInfo(provider.IntegrationInfo{Name: "open_policy_agent", Key: key}, provider.ApplicationInfo{})
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
	service := openpolicyagent.OpaService{}
	p := openpolicyagent.OpaProvider{Client: client, Service: service}
	_, err := p.GetPolicyInfo(provider.IntegrationInfo{Name: "open_policy_agent", Key: key}, provider.ApplicationInfo{})
	assert.Error(t, err)
}

func TestSetPolicyInfo(t *testing.T) {
	key := []byte(`
{
  "bundle_url": "aBigUrl"
}
`)
	mockClient := openpolicyagent_test.MockClient{}
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := openpolicyagent.OpaService{ResourcesDirectory: resourcesDirectory}

	p := openpolicyagent.OpaProvider{Client: client, Service: service}
	err := p.SetPolicyInfo(
		provider.IntegrationInfo{Name: "open_policy_agent", Key: key},
		provider.ApplicationInfo{},
		provider.PolicyInfo{Version: "0.1", Action: "GET", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"allusers"}}, Object: provider.ObjectInfo{Resources: []string{"/"}}},
	)
	assert.NoError(t, err)
}

func TestMakeDefaultBundle(t *testing.T) {
	client := openpolicyagent.BundleClient{}
	service := openpolicyagent.OpaService{}
	p := openpolicyagent.OpaProvider{Client: client, Service: service}

	rego := []byte(`package authz
import future.keywords.in
default allow = false
allow {
	input.method = "GET"
	input.path in ["/"]
	input.principals[_] in ["allusers"]
}`)
	bundle, _ := p.MakeDefaultBundle(rego)

	gzip, _ := compressionsupport.UnGzip(bytes.NewReader(bundle.Bytes()))
	path := filepath.Join(os.TempDir(), fmt.Sprintf("/test-bundle-%d", rand.Uint64()))
	_ = compressionsupport.UnTarToPath(bytes.NewReader(gzip), path)

	created, _ := ioutil.ReadFile(filepath.Join(path, "/bundle/policy.rego"))
	assert.Contains(t, string(created), "input.principals[_] in [\"allusers\"]")
	assert.NotContains(t, string(created), "input.principals[_] in [\"humanresources@\"]")

	mcreated, _ := ioutil.ReadFile(filepath.Join(path, "/bundle/.manifest"))
	assert.Contains(t, string(mcreated), "{\"revision\":\"\",\"roots\":[\"\"]}")

	dcreated, _ := ioutil.ReadFile(filepath.Join(path, "/bundle/data.json"))
	assert.Contains(t, string(dcreated), "{}")
}
