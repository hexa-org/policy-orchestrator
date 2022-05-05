package decisionproviders_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupport/decisionproviders"
	"testing"
)

func TestDefaultProvider_BuildInput(t *testing.T) {
	provider := decisionproviders.DefaultProvider{}
	defer func() {
		if recover() == nil {
			t.Fail()
		}
	}()
	_, _ = provider.BuildInput(nil)
}

func TestDefaultProvider_Allow(t *testing.T) {
	provider := decisionproviders.DefaultProvider{}
	defer func() {
		if recover() == nil {
			t.Fail()
		}
	}()
	_, _ = provider.Allow(nil)
}
