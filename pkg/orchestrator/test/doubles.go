package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
)

type NoopProvider struct {
	Discovered int
	Err        error
}

func (n *NoopProvider) Name() string {
	return "noop"
}

func (n *NoopProvider) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo, err error) {
	if info.Name == n.Name() {
		found := []provider.ApplicationInfo{{ObjectID: "anId", Name: "appEngine"}, {ObjectID: "anotherId", Name: "cloudRun"}, {ObjectID: "andAnotherId", Name: "kubernetes"}}
		apps = append(apps, found...)
		n.Discovered = n.Discovered + 3
	}
	return apps, n.Err
}

func (n *NoopProvider) GetPolicyInfo(_ provider.IntegrationInfo, _ provider.ApplicationInfo) ([]provider.PolicyInfo, error) {
	return []provider.PolicyInfo{
		{"aVersion", "anAction", provider.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, provider.ObjectInfo{Resources: []string{"/"}}},
		{"aVersion", "anotherAction", provider.SubjectInfo{AuthenticatedUsers: []string{"anotherUser"}}, provider.ObjectInfo{Resources: []string{"/"}}},
	}, n.Err
}

func (n *NoopProvider) SetPolicyInfo(_ provider.IntegrationInfo, _ provider.ApplicationInfo, _ []provider.PolicyInfo) error {
	return n.Err
}
