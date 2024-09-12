package orchestratorNoopProvider

import (
	"net/http"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
)

type NoopProvider struct {
	OverrideName string
	Discovered   int
	Err          error
}

func (n *NoopProvider) SetTestErr(err error) {
	n.Err = err
}

func (n *NoopProvider) Name() string {
	if n.OverrideName != "" {
		return n.OverrideName
	}
	return "noop"
}

func (n *NoopProvider) DiscoverApplications(info policyprovider.IntegrationInfo) (apps []policyprovider.ApplicationInfo, err error) {
	if info.Name == n.Name() {
		found := []policyprovider.ApplicationInfo{{ObjectID: "anId", Name: "appEngine"}, {ObjectID: "anotherId", Name: "cloudRun"}, {ObjectID: "andAnotherId", Name: "kubernetes"}}
		apps = append(apps, found...)
		n.Discovered = n.Discovered + 3
	}
	return apps, n.Err
}

func (n *NoopProvider) GetPolicyInfo(_ policyprovider.IntegrationInfo, _ policyprovider.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {

	return []hexapolicy.PolicyInfo{
		{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{"anAction"}, Subjects: []string{"user:aUser"}, Object: "anId"},
		{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{"anotherAction"}, Subjects: []string{"user:anotherUser"}, Object: "anId"},
	}, n.Err
}

func (n *NoopProvider) SetPolicyInfo(_ policyprovider.IntegrationInfo, _ policyprovider.ApplicationInfo, _ []hexapolicy.PolicyInfo) (int, error) {
	return http.StatusCreated, n.Err
}
