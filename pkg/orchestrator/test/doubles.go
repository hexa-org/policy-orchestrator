package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/identityquerylanguage"
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

func (n *NoopProvider) GetPolicyInfo(_ provider.IntegrationInfo, _ provider.ApplicationInfo) ([]identityquerylanguage.PolicyInfo, error) {
	return []identityquerylanguage.PolicyInfo{
		{"aVersion", "anAction", identityquerylanguage.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, identityquerylanguage.ObjectInfo{Resources: []string{"/"}}},
		{"aVersion", "anotherAction", identityquerylanguage.SubjectInfo{AuthenticatedUsers: []string{"anotherUser"}}, identityquerylanguage.ObjectInfo{Resources: []string{"/"}}},
	}, n.Err
}

func (n *NoopProvider) SetPolicyInfo(_ provider.IntegrationInfo, _ provider.ApplicationInfo, _ []identityquerylanguage.PolicyInfo) error {
	return n.Err
}
