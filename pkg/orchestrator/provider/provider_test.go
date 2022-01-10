package provider_test

import (
	"github.com/stretchr/testify/assert"
	"hexa/pkg/orchestrator/provider"
	"hexa/pkg/orchestrator/test"
	"testing"
)

func TestDiscovery(t *testing.T) {
	providers := []provider.Provider{&orchestrator_test.NoopDiscovery{}}
	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "noop", Key: []byte("aKey")}
		assert.Equal(t, 3, len(p.DiscoveryApplications(info)))
	}
}
