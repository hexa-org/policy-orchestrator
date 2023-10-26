package policystore

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
)

type PolicyBackendSvc[R any] interface {
	GetPolicies(info idp.AppInfo) ([]rar.ResourceActionRoles, error)
	SetPolicy(rar rar.ResourceActionRoles) error
}
