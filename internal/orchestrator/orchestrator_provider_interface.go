package orchestrator

import (
	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
)

type Provider interface {
	Name() string
	DiscoverApplications(IntegrationInfo) ([]ApplicationInfo, error)
	GetPolicyInfo(IntegrationInfo, ApplicationInfo) ([]hexapolicy.PolicyInfo, error)
	SetPolicyInfo(IntegrationInfo, ApplicationInfo, []hexapolicy.PolicyInfo) (status int, foundErr error)
}

type IntegrationInfo struct {
	Name string
	Key  []byte
}

type ApplicationInfo struct {
	ObjectID    string `validate:"required"`
	Name        string
	Description string
	Service     string
}
