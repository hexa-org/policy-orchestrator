package providers_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupport/providers"
	"testing"
)

func TestOpaDecisionProvider_BuildInput_BuildInput(t *testing.T) {
	provider := providers.OpaDecisionProvider{}
	defer func() {
		if recover() == nil {
			t.Fail()
		}
	}()
	_, _ = provider.BuildInput(nil)
}

func TestOpaDecisionProvider_Allow(t *testing.T) {
	provider := providers.OpaDecisionProvider{}
	defer func() {
		if recover() == nil {
			t.Fail()
		}
	}()
	_, _ = provider.Allow(nil)
}
