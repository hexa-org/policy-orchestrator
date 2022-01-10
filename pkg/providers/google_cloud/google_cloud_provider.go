package google_cloud

import (
	"hexa/pkg/orchestrator/provider"
	"strings"
)

type GoogleProvider struct {
}

func (g GoogleProvider) Name() string {
	return "google cloud"
}

func (g GoogleProvider) DiscoveryApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo) {
	if strings.EqualFold(info.Name, g.Name()) {
		found := []provider.ApplicationInfo{{"appEngine"}, {"cloudRun"}, {"kubernetes"}}
		apps = append(apps, found...)
	}
	return apps
}
