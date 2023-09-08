package orchestrator_test

import (
	"net/http"

	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
)

type NoopProvider struct {
	OverrideName string
	Discovered   int
	Err          error
}

func (n *NoopProvider) Name() string {
	if n.OverrideName != "" {
		return n.OverrideName
	}
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

func (n *NoopProvider) GetPolicyInfo(_ orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {

	return []hexapolicy.PolicyInfo{
		{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"anAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"user:aUser"}}, Object: hexapolicy.ObjectInfo{
			ResourceID: "anId",
		}},
		{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"anotherAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"user:anotherUser"}}, Object: hexapolicy.ObjectInfo{
			ResourceID: "anId",
		}},
	}, n.Err
}

func (n *NoopProvider) SetPolicyInfo(_ orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo, _ []hexapolicy.PolicyInfo) (int, error) {
	return http.StatusCreated, n.Err
}
