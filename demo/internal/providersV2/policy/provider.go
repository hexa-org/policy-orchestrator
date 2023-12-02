package policy

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
)

type PolicyStore[R rar.ResourceActionRolesMapper] interface {
	Provider() (policystore.PolicyBackendSvc[R], error)
}
