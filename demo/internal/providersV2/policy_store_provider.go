package providersV2

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
)

type PolicyStore[R any] interface {
	Provider() (policystore.PolicyBackendSvc[R], error)
}

type EmptyPolicyStore[R any] struct {
}

func NewEmptyPolicyStore() PolicyStore[any] {
	return &EmptyPolicyStore[any]{}
}

func (a *EmptyPolicyStore[R]) Provider() (policystore.PolicyBackendSvc[R], error) {
	return nil, nil
}
