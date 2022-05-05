package orchestrator

import "github.com/hexa-org/policy-orchestrator/pkg/policysupport"

type Provider interface {
	Name() string
	DiscoverApplications(IntegrationInfo) ([]ApplicationInfo, error)
	GetPolicyInfo(IntegrationInfo, ApplicationInfo) ([]policysupport.PolicyInfo, error)
	SetPolicyInfo(IntegrationInfo, ApplicationInfo, []policysupport.PolicyInfo) error
}

type IntegrationInfo struct {
	Name string
	Key  []byte
}

type ApplicationInfo struct {
	ObjectID    string
	Name        string
	Description string
}
