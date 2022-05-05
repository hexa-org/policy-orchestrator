package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
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

func (n *NoopProvider) GetPolicyInfo(_ orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	return []policysupport.PolicyInfo{
		{"aVersion", "anAction", policysupport.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, policysupport.ObjectInfo{Resources: []string{"/"}}},
		{"aVersion", "anotherAction", policysupport.SubjectInfo{AuthenticatedUsers: []string{"anotherUser"}}, policysupport.ObjectInfo{Resources: []string{"/"}}},
	}, n.Err
}

func (n *NoopProvider) SetPolicyInfo(_ orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo, _ []policysupport.PolicyInfo) error {
	return n.Err
}
