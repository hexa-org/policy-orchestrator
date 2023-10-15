package policyprovider

import (
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/v2/core/idp"
)

// ResourceActionRoles
// TODO - Rename to something better
type ResourceActionRoles struct {
	Action   string // http method e.g GET
	Resource string
	Roles    []string
}

type ProviderService interface {
	DiscoverApplications() ([]idp.AppInfo, error)
	GetPolicyInfo(idp.AppInfo) ([]hexapolicy.PolicyInfo, error)
	SetPolicyInfo(idp.AppInfo, []hexapolicy.PolicyInfo) error
}

func Hello(name string) string {
	return "Hello " + name
}
