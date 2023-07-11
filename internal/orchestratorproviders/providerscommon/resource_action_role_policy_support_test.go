package providerscommon_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"testing"
)

func TestCalcResourceActionRolesForUpdate_NoUpdates(t *testing.T) {
	act := providerscommon.CalcResourceActionRolesForUpdate(nil, nil)
	assert.Empty(t, act)

	act = providerscommon.CalcResourceActionRolesForUpdate(
		[]providerscommon.ResourceActionRoles{},
		[]policysupport.PolicyInfo{})
	assert.Empty(t, act)

	expRoles := []string{"some-role-to-add"}
	tmpMap := map[string][]string{policytestsupport.ActionGetProfile: expRoles}
	inputPolicies := policytestsupport.MakeRoleSubjectTestPolicies(tmpMap)
	act = providerscommon.CalcResourceActionRolesForUpdate(
		[]providerscommon.ResourceActionRoles{},
		inputPolicies)
	assert.Empty(t, act)
}

func TestCalcResourceActionRolesForUpdate_RemoveOneAddOne(t *testing.T) {
	tmpMap := map[string][]string{policytestsupport.ActionGetHrUs: {"some-role-to-remove"}}
	existingRars := policytestsupport.MakeRarList(tmpMap)

	expRoles := []string{"some-role-to-add"}
	tmpMap = map[string][]string{policytestsupport.ActionGetHrUs: expRoles}
	inputPolicies := policytestsupport.MakeRoleSubjectTestPolicies(tmpMap)

	actRars := providerscommon.CalcResourceActionRolesForUpdate(existingRars, inputPolicies)
	assert.Equal(t, 1, len(actRars))
	assert.Equal(t, "GET", actRars[0].Action)
	assert.Equal(t, policytestsupport.ResourceHrUs, actRars[0].Resource)
	assert.Equal(t, expRoles, actRars[0].Roles)
}

func TestCalcResourceActionRolesForUpdate_RemoveOneAddOne_MulResources(t *testing.T) {
	expResActions := []string{policytestsupport.ActionGetProfile, policytestsupport.ActionGetHrUs}
	expRes := []string{policytestsupport.ResourceProfile, policytestsupport.ResourceHrUs}
	tmpMap := map[string][]string{
		expResActions[0]: {"some-role-to-remove"},
		expResActions[1]: {"some-role-to-remove"},
	}
	existingRars := policytestsupport.MakeRarList(tmpMap)

	expRoles := []string{"some-role-to-add"}
	tmpMap = map[string][]string{
		expResActions[0]: expRoles,
		expResActions[1]: expRoles,
	}

	inputPolicies := policytestsupport.MakeRoleSubjectTestPolicies(tmpMap)

	actRars := providerscommon.CalcResourceActionRolesForUpdate(existingRars, inputPolicies)
	assert.Equal(t, 2, len(actRars))
	for a, aRar := range actRars {
		assert.Equal(t, "GET", aRar.Action)
		assert.Equal(t, expRes[a], aRar.Resource)
		assert.Equal(t, expRoles, aRar.Roles)
	}
}

func TestCalcResourceActionRolesForUpdate_OnlyAddNoRemove(t *testing.T) {
	tmpMap := map[string][]string{policytestsupport.ActionGetHrUs: {"role2", "role1"}}
	existingRars := policytestsupport.MakeRarList(tmpMap)

	expRoles := []string{"role1", "role3", "role2"}
	tmpMap = map[string][]string{policytestsupport.ActionGetHrUs: expRoles}
	inputPolicies := policytestsupport.MakeRoleSubjectTestPolicies(tmpMap)

	actRars := providerscommon.CalcResourceActionRolesForUpdate(existingRars, inputPolicies)
	assert.Equal(t, 1, len(actRars))
	assert.Equal(t, "GET", actRars[0].Action)
	assert.Equal(t, policytestsupport.ResourceHrUs, actRars[0].Resource)
	slices.Sort(expRoles)
	assert.Equal(t, expRoles, actRars[0].Roles)
}

// func TestCalcResourceActionRolesForUpdate_SkipResourceNotInInput(t *testing.T) { validates
// no updates are found when no input resources match existing
func TestCalcResourceActionRolesForUpdate_SkipResourceNotInInput(t *testing.T) {
	tmpMap := map[string][]string{policytestsupport.ActionGetHrUs: {"role2", "role1"}}
	existingRars := policytestsupport.MakeRarList(tmpMap)

	expRoles := []string{"role1", "role2"}
	tmpMap = map[string][]string{
		policytestsupport.ActionGetHrUs:    expRoles,
		policytestsupport.ActionGetProfile: expRoles,
	}
	inputPolicies := policytestsupport.MakeRoleSubjectTestPolicies(tmpMap)

	actRars := providerscommon.CalcResourceActionRolesForUpdate(existingRars, inputPolicies)
	assert.Equal(t, 0, len(actRars))
}

// TestCalcResourceActionRolesForUpdate_OnlyUpdateResourceFromInput validates
// only an input resource matching existing one is updated
// any input resources that do not match existing are ignored
func TestCalcResourceActionRolesForUpdate_OnlyUpdateResourceFromInput(t *testing.T) {
	tmpMap := map[string][]string{policytestsupport.ActionGetHrUs: {"role2", "role1"}}
	existingRars := policytestsupport.MakeRarList(tmpMap)

	expRoles := []string{"role1", "role3", "role2"}
	tmpMap = map[string][]string{
		policytestsupport.ActionGetHrUs:    expRoles,
		policytestsupport.ActionGetProfile: expRoles,
	}
	inputPolicies := policytestsupport.MakeRoleSubjectTestPolicies(tmpMap)

	actRars := providerscommon.CalcResourceActionRolesForUpdate(existingRars, inputPolicies)
	assert.Equal(t, 1, len(actRars))
	assert.Equal(t, "GET", actRars[0].Action)
	assert.Equal(t, policytestsupport.ResourceHrUs, actRars[0].Resource)
	slices.Sort(expRoles)
	assert.Equal(t, expRoles, actRars[0].Roles)
}

func TestCalcResourceActionRolesForUpdate_RemoveAllRoleAssignments(t *testing.T) {
	tmpMap := map[string][]string{policytestsupport.ActionGetHrUs: {"role2", "role1"}}
	existingRars := policytestsupport.MakeRarList(tmpMap)

	tmpMap = map[string][]string{policytestsupport.ActionGetHrUs: {}}
	inputPolicies := policytestsupport.MakeRoleSubjectTestPolicies(tmpMap)
	actRars := providerscommon.CalcResourceActionRolesForUpdate(existingRars, inputPolicies)
	assert.Equal(t, 1, len(actRars))
	assert.Equal(t, "GET", actRars[0].Action)
	assert.Equal(t, policytestsupport.ResourceHrUs, actRars[0].Resource)
	assert.Equal(t, 0, len(actRars[0].Roles))
}
