package policyprovider

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/internal/testhelper"
	"github.com/hexa-org/policy-orchestrator/sdk/core/internal/testhelper/idql"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestMapIdqlToRar_ErrorWithoutResource(t *testing.T) {
	expMethods := []string{http.MethodGet, http.MethodPost}
	expMembers := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol := idql.MakeTestPolicy("", expMethods, expMembers)
	actRarMap, err := mapIdqlToRar(hexaPol)
	assert.ErrorContains(t, err, "empty resource")
	assert.Nil(t, actRarMap)
}

// TestMapIdqlToRar_ActionTrimmed asserts no error returned if resource has leading/trailing spaces
func TestMapIdqlToRar_ResourceTrimmed(t *testing.T) {
	expResource := testhelper.ResourceHrUs
	expMethods := []string{http.MethodGet, http.MethodPost}
	expMembers := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol := idql.MakeTestPolicy("   "+expResource+"   ", expMethods, expMembers)
	actRarMap, err := mapIdqlToRar(hexaPol)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actRarMap))
	for _, actRar := range actRarMap {
		assert.Equal(t, expResource, actRar.Resource())
	}
}

func TestMapIdqlToRar_InvalidActionsError(t *testing.T) {
	expResource := testhelper.ResourceHrUs
	expMembers := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol := idql.MakeTestPolicy(expResource, nil, expMembers)
	actRarMap, err := mapIdqlToRar(hexaPol)
	assert.ErrorContains(t, err, "nil actionUri")
	assert.Nil(t, actRarMap)

	expMethods := []string{http.MethodGet, ""}
	expMembers = []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol = idql.MakeTestPolicy(expResource, expMethods, expMembers)
	actRarMap, err = mapIdqlToRar(hexaPol)
	assert.ErrorContains(t, err, "without actionUri")
	assert.Nil(t, actRarMap)

	expMethods = []string{http.MethodGet, "http:SOMETHING"}
	expMembers = []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol = idql.MakeTestPolicy(expResource, expMethods, expMembers)
	actRarMap, err = mapIdqlToRar(hexaPol)
	assert.ErrorContains(t, err, "Invalid http method")
	assert.Nil(t, actRarMap)

	//expMethods = []string{http.MethodGet, "  "}
	//expMembers = []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	//hexaPol = idql.MakeTestPolicy(expResource, expMethods, expMembers)
	//actRarMap, err = mapIdqlToRar(hexaPol)
	//assert.ErrorContains(t, err, "Invalid http method")
	//assert.Nil(t, actRarMap)
}

// TestMapIdqlToRar_ActionTrimmed asserts no error returned if action has leading/trailing spaces
func TestMapIdqlToRar_ActionTrimmed(t *testing.T) {
	expResource := testhelper.ResourceHrUs
	expMethods := []string{http.MethodGet + "  "}
	expMembers := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol := idql.MakeTestPolicy(expResource, expMethods, expMembers)
	actRarMap, err := mapIdqlToRar(hexaPol)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actRarMap))
	expRarMap := testhelper.MakeRarMap(expResource, []string{http.MethodGet}, expMembers)
	assert.Equal(t, expRarMap, actRarMap)
}

// TestMapIdqlToRar - assets that when given idql with multiple actions
// mapIdqlToRar maps each action to a separate rar
func TestMapIdqlToRar_IdqlWithMultipleActions(t *testing.T) {
	// Two hexa policies with same resource action, having different members
	// One hexa policy with multiple actions, multiple members
	expResource := testhelper.ResourceHrUs
	expMethods := []string{http.MethodGet, http.MethodPost}
	expMembers := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol := idql.MakeTestPolicy(expResource, expMethods, expMembers)

	actRarMap, err := mapIdqlToRar(hexaPol)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actRarMap))
	expRarMap := testhelper.MakeRarMap(expResource, expMethods, expMembers)
	assert.Equal(t, expRarMap, actRarMap)
}

// TestMapIdqlToRar - assets that when given multiple idql policies with same resource, actions but different members
// mapIdqlToRar maps merges the members into a single rar
func TestMapIdqlToRar_MembersMergedInRar(t *testing.T) {
	expResource := testhelper.ResourceHrUs
	expMethods := []string{http.MethodGet, http.MethodPost}
	expMembers1 := []string{testhelper.RoleReadHrUs, "  ", ""}
	hexaPol1 := idql.MakeTestPolicy(expResource, expMethods, expMembers1)

	expMembers2 := []string{"  ", "", testhelper.RoleReadProfile}
	hexaPol2 := idql.MakeTestPolicy(expResource, expMethods, expMembers2)

	actRarMap, err := mapIdqlToRar(hexaPol1, hexaPol2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actRarMap))
	expMembers := make([]string, 0)
	expMembers = append(expMembers, expMembers1...)
	expMembers = append(expMembers, expMembers2...)
	expRarMap := testhelper.MakeRarMap(expResource, expMethods, expMembers)
	assert.Equal(t, expRarMap, actRarMap)
}

// TestMapIdqlToRar_DuplicateIdqlPolicies asserts that when given duplicate idql policies
// mapIdqlToRar removes the duplicates
func TestMapIdqlToRar_DuplicateIdqlPolicies(t *testing.T) {
	expResource := testhelper.ResourceHrUs
	expMethods := []string{http.MethodGet, http.MethodPost}
	expMembers := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol1 := idql.MakeTestPolicy(expResource, expMethods, expMembers)
	hexaPol2 := idql.MakeTestPolicy(expResource, expMethods, expMembers)

	actRarMap, err := mapIdqlToRar(hexaPol1, hexaPol2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actRarMap))

	expRarMap := testhelper.MakeRarMap(expResource, expMethods, expMembers)
	assert.Equal(t, expRarMap, actRarMap)
}

// TestMapIdqlToRar_NoMembers asserts no error returned if no members in IDQL
func TestMapIdqlToRar_NoMembers(t *testing.T) {
	expResource := testhelper.ResourceHrUs
	expMethods := []string{http.MethodGet, http.MethodPost}
	tests := []struct {
		name    string
		members []string
	}{
		{name: "Nil members", members: nil},
		{name: "No members", members: []string{}},
		{name: "Empty members", members: []string{"", "  "}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hexaPol1 := idql.MakeTestPolicy(expResource, expMethods, tt.members)
			actRarMap, err := mapIdqlToRar(hexaPol1)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(actRarMap))
			expRarMap := testhelper.MakeRarMap(expResource, expMethods, tt.members)
			assert.Equal(t, expRarMap, actRarMap)
		})
	}
}

// TestMapIdqlToRar_MultiplePoliciesWithMultipleActions asserts a valid map
// is returned when multiple idql policies are provided
func TestMapIdqlToRar_MultiplePoliciesWithMultipleActions(t *testing.T) {
	expResource1 := testhelper.ResourceHrUs
	expMethods1 := []string{http.MethodGet, http.MethodPost}
	expMembers1 := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol1 := idql.MakeTestPolicy(expResource1, expMethods1, expMembers1)

	expResource2 := testhelper.ResourceProfile
	expMethods2 := []string{http.MethodPut, http.MethodDelete}
	expMembers2 := []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}
	hexaPol2 := idql.MakeTestPolicy(expResource2, expMethods2, expMembers2)

	actRarMap, err := mapIdqlToRar(hexaPol1, hexaPol2)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(actRarMap))
	expRarMap := testhelper.MakeRarMapMultiple([]string{expResource1, expResource2}, [][]string{expMethods1, expMethods2}, [][]string{expMembers1, expMembers2})

	assert.Equal(t, expRarMap, actRarMap)
}
