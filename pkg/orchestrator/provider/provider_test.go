package provider_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiscovery(t *testing.T) {
	providers := []provider.Provider{&orchestrator_test.NoopDiscovery{}}
	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "noop", Key: []byte("aKey")}
		applications, _ := p.DiscoverApplications(info)
		assert.Equal(t, 3, len(applications))
	}
}
