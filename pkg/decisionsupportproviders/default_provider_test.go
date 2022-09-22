package decisionsupportproviders_test

import (
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupportproviders"
)

func TestDefaultProvider_BuildInput(t *testing.T) {
	provider := decisionsupportproviders.DefaultProvider{}
	defer func() {
		if recover() == nil {
			t.Fail()
		}
	}()
	_, _ = provider.BuildInput(nil)
}

func TestDefaultProvider_Allow(t *testing.T) {
	provider := decisionsupportproviders.DefaultProvider{}
	defer func() {
		if recover() == nil {
			t.Fail()
		}
	}()
	_, _ = provider.Allow(nil)
}
