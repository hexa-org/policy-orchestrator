package decisionsupportproviders_test

import (
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupportproviders"
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
