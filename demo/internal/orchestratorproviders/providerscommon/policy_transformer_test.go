package providerscommon_test

import (
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildPolicies_EmptyNil(t *testing.T) {
	policies := providerscommon.BuildPolicies(nil)
	assert.NotNil(t, policies)
	assert.Empty(t, policies)

	policies = providerscommon.BuildPolicies([]providerscommon.ResourceActionRoles{})
	assert.NotNil(t, policies)
	assert.Empty(t, policies)
}

func TestBuildPolicies(t *testing.T) {
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetProfile: {"2-some-profile-role", "1-some-profile-role"},
		policytestsupport.ActionGetHrUs:    {"2-some-hr-role", "1-some-hr-role"},
	}
	resourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policies := providerscommon.BuildPolicies(resourceRoles)
	assert.NotNil(t, policies)
	assert.Len(t, policies, 2)

	actPol := policies[0]
	assert.Equal(t, policytestsupport.ResourceHrUs, actPol.Object.ResourceID)
	assert.Equal(t, "http:GET", actPol.Actions[0].ActionUri)
	assert.Equal(t, []string{"1-some-hr-role", "2-some-hr-role"}, actPol.Subject.Members)

	actPol = policies[1]
	assert.Equal(t, policytestsupport.ResourceProfile, actPol.Object.ResourceID)
	assert.Equal(t, "http:GET", actPol.Actions[0].ActionUri)
	assert.Equal(t, []string{"1-some-profile-role", "2-some-profile-role"}, actPol.Subject.Members)
}

func TestCompactActions_NilEmpty(t *testing.T) {
	tests := []struct {
		name     string
		existing []hexapolicy.ActionInfo
		newOnes  []hexapolicy.ActionInfo
	}{
		{name: "nils", existing: nil, newOnes: nil},
		{name: "empties", existing: []hexapolicy.ActionInfo{}, newOnes: []hexapolicy.ActionInfo{}},
		{name: "existing nil", existing: nil, newOnes: []hexapolicy.ActionInfo{}},
		{name: "newOnes nil", existing: []hexapolicy.ActionInfo{}, newOnes: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compacted := providerscommon.CompactActions(tt.existing, tt.newOnes)
			assert.NotNil(t, compacted)
			assert.Empty(t, compacted)
		})
	}
}

func TestCompactActions_AllWhitespace(t *testing.T) {
	arr1 := []hexapolicy.ActionInfo{
		{ActionUri: ""}, {ActionUri: "   "}, {ActionUri: " "},
	}
	compacted := providerscommon.CompactActions(arr1, arr1)
	assert.NotNil(t, compacted)
	assert.Empty(t, compacted)
}

func TestCompactActions_DuplicatesAndWhitespace(t *testing.T) {
	arr1 := []hexapolicy.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}
	arr2 := []hexapolicy.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}

	compacted := providerscommon.CompactActions(arr1, arr2)
	assert.NotNil(t, compacted)
	assert.Equal(t, []hexapolicy.ActionInfo{
		{ActionUri: "1one"}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}, compacted)
}

func TestCompactActions_UniqueAndWhitespace(t *testing.T) {
	arr1 := []hexapolicy.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}
	arr2 := []hexapolicy.ActionInfo{
		{ActionUri: ""}, {ActionUri: "4four"}, {ActionUri: " "}, {ActionUri: "5five"},
	}

	compacted := providerscommon.CompactActions(arr1, arr2)
	assert.NotNil(t, compacted)
	assert.Equal(t, []hexapolicy.ActionInfo{
		{ActionUri: "1one"}, {ActionUri: "2two"}, {ActionUri: "3three"}, {ActionUri: "4four"}, {ActionUri: "5five"},
	}, compacted)
}

func TestCompactActions_OneEmptyNil(t *testing.T) {
	arr := []hexapolicy.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}

	compacted := providerscommon.CompactActions(arr, nil)
	assert.NotNil(t, compacted)
	assert.Equal(t, []hexapolicy.ActionInfo{
		{ActionUri: "1one"}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}, compacted)

	compacted = providerscommon.CompactActions(nil, arr)
	assert.NotNil(t, compacted)
	assert.Equal(t, []hexapolicy.ActionInfo{
		{ActionUri: "1one"}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}, compacted)
}

func TestCompactMembers_Nil(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		newOnes  []string
	}{
		{name: "nils", existing: nil, newOnes: nil},
		{name: "empties", existing: []string{}, newOnes: []string{}},
		{name: "existing nil, newOnes empty", existing: nil, newOnes: []string{}},
		{name: "existing empty, newOnes nil", existing: []string{}, newOnes: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compacted := providerscommon.CompactMembers(tt.existing, tt.newOnes)
			assert.NotNil(t, compacted)
			assert.Empty(t, compacted)
		})
	}
}

func TestCompactMembers_AllWhitespace(t *testing.T) {
	compacted := providerscommon.CompactMembers([]string{"", "", " ", "  ", "", " "}, []string{"", "", " ", "  ", "", " "})
	assert.NotNil(t, compacted)
	assert.Empty(t, compacted)
}

func TestCompactMembers_DuplicatesAndWhitespace(t *testing.T) {
	arr := []string{"hello", "", "how", "are", " ", "you", "hello", "   ", "hello", "", "how", "are", "you", " "}
	compacted := providerscommon.CompactMembers(arr, arr)
	assert.Equal(t, []string{"are", "hello", "how", "you"}, compacted)
}

func TestCompactMembers_UniqueWhitespace(t *testing.T) {
	arr1 := []string{"hello", "", "how", "are", " ", "you"}
	arr2 := []string{"i", "", "am", "find", " ", "thank", "you"}
	compacted := providerscommon.CompactMembers(arr1, arr2)
	assert.Equal(t, []string{"am", "are", "find", "hello", "how", "i", "thank", "you"}, compacted)
}

func TestCompactMembers_OneNil(t *testing.T) {
	arr := []string{"1one", "", "2two", "3three", " ", "4four"}
	compacted := providerscommon.CompactMembers(arr, nil)
	assert.Equal(t, []string{"1one", "2two", "3three", "4four"}, compacted)

	compacted = providerscommon.CompactMembers(nil, arr)
	assert.Equal(t, []string{"1one", "2two", "3three", "4four"}, compacted)
}

func TestFlattenPolicy_ReturnsEmpty(t *testing.T) {
	actPolicies := providerscommon.FlattenPolicy([]hexapolicy.PolicyInfo{})
	assert.NotNil(t, actPolicies)
	assert.Empty(t, actPolicies)

	actPolicies = providerscommon.FlattenPolicy(nil)
	assert.NotNil(t, actPolicies)
	assert.Empty(t, actPolicies)
}

func TestFlattenPolicy_DupResourceDupMembers(t *testing.T) {
	pol1 := hexapolicy.PolicyInfo{
		Meta: hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{
			{ActionUri: ""}, {ActionUri: "1act"}, {ActionUri: " "}, {ActionUri: "2act"}},
		Subject: hexapolicy.SubjectInfo{Members: []string{"1mem", "", "2mem"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource1"},
	}

	pol2 := hexapolicy.PolicyInfo{
		Meta: hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{
			{ActionUri: ""}, {ActionUri: "3act"}, {ActionUri: " "}, {ActionUri: "4act"}},
		Subject: hexapolicy.SubjectInfo{Members: []string{"1mem", "", "2mem"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource1"},
	}

	orig := []hexapolicy.PolicyInfo{pol1, pol2}
	actPolicies := providerscommon.FlattenPolicy(orig)
	assert.NotNil(t, actPolicies)
	assert.Equal(t, 4, len(actPolicies))

	expResource := "resource1"
	expActions := []string{"1act", "2act", "3act", "4act"}
	expMembers := []string{"1mem", "2mem"}

	for i, actPol := range actPolicies {
		assert.Equal(t, expResource, actPol.Object.ResourceID)
		assert.Equal(t, expActions[i], actPol.Actions[0].ActionUri)
		assert.Equal(t, expMembers, actPol.Subject.Members)
	}
}

func TestFlattenPolicy_NoResource(t *testing.T) {
	pol1 := hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{{ActionUri: "1act"}, {ActionUri: "2act"}},
		Subject: hexapolicy.SubjectInfo{Members: []string{"1mem", "", "2mem"}},
	}
	pol2 := hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{{ActionUri: "1act"}},
		Subject: hexapolicy.SubjectInfo{Members: []string{"1mem", "2mem"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource1"},
	}

	tests := []struct {
		name          string
		inputPolicies []hexapolicy.PolicyInfo
		expLen        int
	}{
		{
			name:          "Single policy without resource",
			inputPolicies: []hexapolicy.PolicyInfo{pol1},
		},
		{
			name:          "Two policies one with, one without resource",
			inputPolicies: []hexapolicy.PolicyInfo{pol1, pol2},
			expLen:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := tt.inputPolicies
			actPolicies := providerscommon.FlattenPolicy(orig)
			assert.NotNil(t, actPolicies)
			assert.Len(t, actPolicies, tt.expLen)
		})
	}

}

func TestFlattenPolicy_NoActions(t *testing.T) {
	pol1 := hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Subject: hexapolicy.SubjectInfo{Members: []string{"1mem", "", "2mem"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource1"},
	}
	orig := []hexapolicy.PolicyInfo{pol1}
	actPolicies := providerscommon.FlattenPolicy(orig)
	assert.NotNil(t, actPolicies)
	assert.Equal(t, []hexapolicy.PolicyInfo{}, actPolicies)
}

func TestFlattenPolicy_NoMembers(t *testing.T) {
	pol1 := hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{{ActionUri: "1act"}, {ActionUri: "2act"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource1"},
	}
	orig := []hexapolicy.PolicyInfo{pol1}
	actPolicies := providerscommon.FlattenPolicy(orig)
	assert.NotNil(t, actPolicies)
	assert.Equal(t, 2, len(actPolicies))

	expActions := []string{"1act", "2act"}
	for i, actPol := range actPolicies {
		assert.Equal(t, "resource1", actPol.Object.ResourceID)
		assert.Equal(t, expActions[i], actPol.Actions[0].ActionUri)
		assert.NotNil(t, actPol.Subject.Members)
		assert.Equal(t, []string{}, actPol.Subject.Members)
	}
}

func TestFlattenPolicy_MergeSameResourceAction(t *testing.T) {
	pol1a := hexapolicy.PolicyInfo{
		Meta: hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{
			{ActionUri: "1act"}, {ActionUri: "2act"}},
		Subject: hexapolicy.SubjectInfo{Members: []string{"1mem", "2mem"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource1"},
	}

	pol1b := hexapolicy.PolicyInfo{
		Meta: hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{
			{ActionUri: "1act"}, {ActionUri: "2act"}},
		Subject: hexapolicy.SubjectInfo{Members: []string{"3mem", "4mem"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource1"},
	}

	pol2 := hexapolicy.PolicyInfo{
		Meta: hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{
			{ActionUri: "3act"}, {ActionUri: "4act"}},
		Subject: hexapolicy.SubjectInfo{Members: []string{"1mem", "2mem"}},
		Object:  hexapolicy.ObjectInfo{ResourceID: "resource2"},
	}

	orig := []hexapolicy.PolicyInfo{pol1a, pol2, pol1b}
	actPolicies := providerscommon.FlattenPolicy(orig)

	assert.NotNil(t, actPolicies)
	assert.Equal(t, 4, len(actPolicies))

	expResource := "resource1"
	expMembers := []string{"1mem", "2mem", "3mem", "4mem"}
	expActions := []string{"1act", "2act"}
	for i := 0; i < len(expActions); i++ {
		actPol := actPolicies[i]
		assert.NotNil(t, actPol)
		assert.Equal(t, expResource, actPol.Object.ResourceID)
		assert.Equal(t, expActions[i], actPol.Actions[0].ActionUri)
		assert.Equal(t, expMembers, actPol.Subject.Members)
	}

	expResource = "resource2"
	expMembers = []string{"1mem", "2mem"}
	expActions = []string{"3act", "4act"}
	for i := 0; i < len(expActions); i++ {
		actPol := actPolicies[i+2]
		assert.NotNil(t, actPol)
		assert.Equal(t, expResource, actPol.Object.ResourceID)
		assert.Equal(t, expActions[i], actPol.Actions[0].ActionUri)
		assert.Equal(t, expMembers, actPol.Subject.Members)
	}
}
