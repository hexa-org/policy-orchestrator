package openpolicyagent_test

import (
	"bytes"
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
