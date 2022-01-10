package google_cloud_test

import (
	"github.com/stretchr/testify/assert"
	"hexa/pkg/orchestrator/provider"
	"hexa/pkg/providers/google_cloud"
	"testing"
)

func TestDiscovery(t *testing.T) {
	providers := []provider.Provider{google_cloud.GoogleProvider{}}
	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "google cloud", Key: []byte("aKey")}
		assert.Equal(t, 3, len(p.DiscoveryApplications(info)))
		assert.Equal(t, "google cloud", p.Name())
	}
}

func TestDiscovery_ignores_case(t *testing.T) {
	providers := []provider.Provider{google_cloud.GoogleProvider{}}
	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "Google Cloud", Key: []byte("aKey")}
		assert.Equal(t, 3, len(p.DiscoveryApplications(info)))
		assert.Equal(t, "google cloud", p.Name())
	}
}

func TestNoDiscovery(t *testing.T) {
	providers := []provider.Provider{google_cloud.GoogleProvider{}}
	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "not google cloud", Key: []byte("aKey")}
		assert.Equal(t, 0, len(p.DiscoveryApplications(info)))
	}
}
