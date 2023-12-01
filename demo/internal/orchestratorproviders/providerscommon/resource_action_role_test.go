package providerscommon_test

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/providerscommon"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const actionUri = "http:GET"

func TestNewResourceActionRoles_Invalid(t *testing.T) {
	act := providerscommon.NewResourceActionRoles("/some", "INVALID", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionRoles("/some", "http:GET", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionRoles("/some", "httpget", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionRoles("  ", http.MethodGet, []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)
}

func TestNewResourceActionRoles_Success(t *testing.T) {
	act := providerscommon.NewResourceActionRoles("/some", http.MethodGet, []string{"mem1", "mem2"})
	assert.Equal(t, providerscommon.ResourceActionRoles{
		Action:   http.MethodGet,
		Resource: "/some",
		Roles:    []string{"mem1", "mem2"},
	}, act)
}

func TestNewResourceActionUriRoles_InvalidMethod(t *testing.T) {
	act := providerscommon.NewResourceActionUriRoles("/some", "INVALID", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionUriRoles("/some", http.MethodGet, []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionUriRoles("/some", "httpget", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)
}

func TestNewResourceActionUriRoles_Success(t *testing.T) {
	act := providerscommon.NewResourceActionUriRoles("/some", "http:GET", []string{"mem1", "mem2"})
	assert.Equal(t, providerscommon.ResourceActionRoles{
		Action:   http.MethodGet,
		Resource: "/some",
		Roles:    []string{"mem1", "mem2"},
	}, act)
}

func TestNewResourceActionRolesFromProviderValue_Invalid(t *testing.T) {
	act := providerscommon.NewResourceActionRolesFromProviderValue("invalid", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionRolesFromProviderValue("badprefix-httpget-some", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionRolesFromProviderValue("resrol-httpbadmethod-some", []string{})
	assert.Equal(t, providerscommon.ResourceActionRoles{}, act)

	act = providerscommon.NewResourceActionRolesFromProviderValue("resrol-httpget-some", []string{"mem1", "mem2"})
	assert.Equal(t, providerscommon.ResourceActionRoles{
		Action:   http.MethodGet,
		Resource: "/some",
		Roles:    []string{"mem1", "mem2"},
	}, act)
}

func TestNewResourceActionRolesFromProviderValue(t *testing.T) {
	resActionKey := "resrol-httpget-humanresources-us"
	act := providerscommon.NewResourceActionRolesFromProviderValue(resActionKey, []string{"some-role"})
	assert.Equal(t, http.MethodGet, act.Action)
	assert.Equal(t, "/humanresources/us", act.Resource)
	assert.Equal(t, []string{"some-role"}, act.Roles)
}

func TestMakeRarKeyForPolicy_Invalid(t *testing.T) {
	aKey := providerscommon.MakeRarKeyForPolicy("  ", "/humanresources/us")
	assert.Equal(t, "", aKey)

	aKey = providerscommon.MakeRarKeyForPolicy(actionUri, "  ")
	assert.Equal(t, "", aKey)

	aKey = providerscommon.MakeRarKeyForPolicy(http.MethodGet, "/humanresources/us")
	assert.Equal(t, "", aKey)
}

func TestMakeRarKeyForPolicy(t *testing.T) {
	aKey := providerscommon.MakeRarKeyForPolicy(actionUri, "/humanresources/us")
	assert.Equal(t, "resrol-httpget-humanresources-us", aKey)
}

func TestNameValue(t *testing.T) {
	resRole := providerscommon.NewResourceActionUriRoles("/humanresources/us", actionUri, []string{"some-role"})
	assert.Equal(t, "resrol-httpget-humanresources-us", resRole.Name())
	assert.Equal(t, `["some-role"]`, resRole.Value())
}
