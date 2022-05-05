package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/identityquerylanguage"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
)

type NoopProvider struct {
	Discovered int
	Err        error
}

func (n *NoopProvider) Name() string {
	return "noop"
}

func (n *NoopProvider) DiscoverApplications(info orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	if info.Name == n.Name() {
		found := []orchestrator.ApplicationInfo{{ObjectID: "anId", Name: "appEngine"}, {ObjectID: "anotherId", Name: "cloudRun"}, {ObjectID: "andAnotherId", Name: "kubernetes"}}
		apps = append(apps, found...)
		n.Discovered = n.Discovered + 3
	}
	return apps, n.Err
}

func (n *NoopProvider) GetPolicyInfo(_ orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo) ([]identityquerylanguage.PolicyInfo, error) {
	return []identityquerylanguage.PolicyInfo{
		{"aVersion", "anAction", identityquerylanguage.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, identityquerylanguage.ObjectInfo{Resources: []string{"/"}}},
		{"aVersion", "anotherAction", identityquerylanguage.SubjectInfo{AuthenticatedUsers: []string{"anotherUser"}}, identityquerylanguage.ObjectInfo{Resources: []string{"/"}}},
	}, n.Err
}

func (n *NoopProvider) SetPolicyInfo(_ orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo, _ []identityquerylanguage.PolicyInfo) error {
	return n.Err
}
