package decisionsupportproviders_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupportproviders"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBuildInput(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{}
	provider.On("BuildInput").Once()

	req, _ := http.NewRequest("GET", "/noop", nil)

	input, _ := provider.BuildInput(req)
	casted := input.(map[string]string)

	assert.Equal(t, "GET", casted["method"])
	assert.Equal(t, "/noop", casted["path"])
	provider.AssertExpectations(t)
}

func TestAllow(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{Decision: true}
	provider.On("BuildInput").Once()
	provider.On("Allow").Once()

	req, _ := http.NewRequest("GET", "/noop", nil)

	input, _ := provider.BuildInput(req)
	allow, _ := provider.Allow(input)

	assert.True(t, allow)
	provider.AssertExpectations(t)
}

func TestAllow_notAllowed(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{Decision: false}
	provider.On("BuildInput").Once()
	provider.On("Allow").Once()

	req, _ := http.NewRequest("GET", "/unauthorized", nil)

	input, _ := provider.BuildInput(req)
	allow, _ := provider.Allow(input)

	assert.False(t, allow)
	provider.AssertExpectations(t)
}
