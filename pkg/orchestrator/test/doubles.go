package orchestrator_test

import (
	"hexa/pkg/orchestrator/provider"
)

type NoopDiscovery struct {
	Discovered int
}

func (n *NoopDiscovery) Name() string {
	return "noop"
}

func (n *NoopDiscovery) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo) {
	if info.Name == n.Name() {
		found := []provider.ApplicationInfo{{ID: "anId", Name: "appEngine"}, {ID: "anotherId", Name: "cloudRun"}, {ID: "andAnotherId", Name: "kubernetes"}}
		apps = append(apps, found...)
		n.Discovered = n.Discovered + 3
	}
	return apps
}
