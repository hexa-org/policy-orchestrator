package providers_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupport/providers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBuildInput(t *testing.T) {
	provider := providers.MockDecisionProvider{}
	provider.On("BuildInput")

	req, _ := http.NewRequest("GET", "/noop", nil)

	input, _ := provider.BuildInput(req)
	casted := input.(map[string]string)

	assert.Equal(t, "GET", casted["method"])
	assert.Equal(t, "/noop", casted["path"])
}

func TestAllow(t *testing.T) {
	provider := providers.MockDecisionProvider{Decision: true}
	provider.On("BuildInput")
	provider.On("Allow")

	req, _ := http.NewRequest("GET", "/noop", nil)

	input, _ := provider.BuildInput(req)
	allow, _ := provider.Allow(input)

	assert.True(t, allow)
}

func TestAllow_notAllowed(t *testing.T) {
	provider := providers.MockDecisionProvider{Decision: false}
	provider.On("BuildInput")
	provider.On("Allow")

	req, _ := http.NewRequest("GET", "/unauthorized", nil)

	input, _ := provider.BuildInput(req)
	allow, _ := provider.Allow(input)

	assert.False(t, allow)
}
