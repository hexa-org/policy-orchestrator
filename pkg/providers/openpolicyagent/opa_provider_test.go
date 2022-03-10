package openpolicyagent_test

import (
	"bytes"
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/openpolicyagent"
	"github.com/stretchr/testify/assert"
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
	applications, _ := p.DiscoverApplications(provider.IntegrationInfo{Name: "open_policy_agent", Key:  key,})
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

	mockClient := MockClient{response: buffer.Bytes()}
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
	mockClient := MockClient{}
	mockClient.err = errors.New("oops")
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
	mockClient := MockClient{}
	mockClient.err = errors.New("oops")
	client := openpolicyagent.BundleClient{HttpClient: &mockClient}
	service := openpolicyagent.OpaService{}
	p := openpolicyagent.OpaProvider{Client: client, Service: service}
	_, err := p.GetPolicyInfo(provider.IntegrationInfo{Name: "open_policy_agent", Key: key}, provider.ApplicationInfo{})
	assert.Error(t, err)
}

func TestSetPolicyInfo(t *testing.T) {
	client := openpolicyagent.BundleClient{}
	service := openpolicyagent.OpaService{}
	p := openpolicyagent.OpaProvider{Client: client, Service: service}
	err := p.SetPolicyInfo(provider.IntegrationInfo{}, provider.ApplicationInfo{}, provider.PolicyInfo{})
	assert.NoError(t, err)
}
