package providerscommon_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const actionUri = "http:GET"

func TestMakeRarKeyForPolicy(t *testing.T) {
	aKey := providerscommon.MakeRarKeyForPolicy(actionUri, "/humanresources/us")
	assert.Equal(t, "resrol-httpget-humanresources-us", aKey)
}

func TestNameValue(t *testing.T) {
	resRole := providerscommon.NewResourceActionRoles("/humanresources/us", actionUri, []string{"some-role"})
	assert.Equal(t, "resrol-httpget-humanresources-us", resRole.Name())
	assert.Equal(t, `["some-role"]`, resRole.Value())
}

func TestNewResourceActionRolesFromProviderValue(t *testing.T) {
	resActionKey := "resrol-httpget-humanresources-us"
	act := providerscommon.NewResourceActionRolesFromProviderValue(resActionKey, []string{"some-role"})
	assert.Equal(t, http.MethodGet, act.Action)
	assert.Equal(t, "/humanresources/us", act.Resource)
	assert.Equal(t, []string{"some-role"}, act.Roles)
}
