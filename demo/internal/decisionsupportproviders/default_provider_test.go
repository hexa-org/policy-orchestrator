package decisionsupportproviders_test

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/decisionsupportproviders"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultProvider_BuildInput(t *testing.T) {
	provider := decisionsupportproviders.DefaultProvider{}
	assert.Panics(t, func() { provider.BuildInput(nil) })
}

func TestDefaultProvider_Allow(t *testing.T) {
	provider := decisionsupportproviders.DefaultProvider{}
	assert.Panics(t, func() { provider.Allow(nil) })
}
