package policystore

import (
	"github.com/hexa-org/policy-orchestrator/v2/core/idp"
	"github.com/hexa-org/policy-orchestrator/v2/core/policyprovider"
)

type PolicyStoreSvc interface {
	GetPolicies(info idp.AppInfo) ([]policyprovider.ResourceActionRoles, error)
	SetPolicy(rar policyprovider.ResourceActionRoles) error
}
