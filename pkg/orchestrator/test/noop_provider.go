package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"net/http"
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
		{policysupport.MetaInfo{Version: "aVersion"}, []policysupport.ActionInfo{{"anAction"}}, policysupport.SubjectInfo{Members: []string{"aUser"}}, policysupport.ObjectInfo{
			ResourceID: "anId",
		}},
		{policysupport.MetaInfo{Version: "aVersion"}, []policysupport.ActionInfo{{"anotherAction"}}, policysupport.SubjectInfo{Members: []string{"anotherUser"}}, policysupport.ObjectInfo{
			ResourceID: "anotherId",
		}},
	}, n.Err
}

func (n *NoopProvider) SetPolicyInfo(_ orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo, _ []policysupport.PolicyInfo) (int, error) {
	return http.StatusCreated, n.Err
}
