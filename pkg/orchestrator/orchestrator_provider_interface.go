package orchestrator

import "github.com/hexa-org/policy-orchestrator/pkg/identityquerylanguage"

type Provider interface {
	Name() string
	DiscoverApplications(IntegrationInfo) ([]ApplicationInfo, error)
	GetPolicyInfo(IntegrationInfo, ApplicationInfo) ([]identityquerylanguage.PolicyInfo, error)
	SetPolicyInfo(IntegrationInfo, ApplicationInfo, []identityquerylanguage.PolicyInfo) error
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
