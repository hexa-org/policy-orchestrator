package orchestrator

import (
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
)

type Provider interface {
	Name() string
	DiscoverApplications(IntegrationInfo) ([]ApplicationInfo, error)
	GetPolicyInfo(IntegrationInfo, ApplicationInfo) ([]policysupport.PolicyInfo, error)
	SetPolicyInfo(IntegrationInfo, ApplicationInfo, []policysupport.PolicyInfo) (status int, foundErr error)
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
