package providerscommon_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompactActions_NilEmpty(t *testing.T) {
	tests := []struct {
		name     string
		existing []policysupport.ActionInfo
		newOnes  []policysupport.ActionInfo
	}{
		{name: "nils", existing: nil, newOnes: nil},
		{name: "empties", existing: []policysupport.ActionInfo{}, newOnes: []policysupport.ActionInfo{}},
		{name: "existing nil", existing: nil, newOnes: []policysupport.ActionInfo{}},
		{name: "newOnes nil", existing: []policysupport.ActionInfo{}, newOnes: nil},
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
	arr1 := []policysupport.ActionInfo{
		{ActionUri: ""}, {ActionUri: "   "}, {ActionUri: " "},
	}
	compacted := providerscommon.CompactActions(arr1, arr1)
	assert.NotNil(t, compacted)
	assert.Empty(t, compacted)
}

func TestCompactActions_DuplicatesAndWhitespace(t *testing.T) {
	arr1 := []policysupport.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}
	arr2 := []policysupport.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}

	compacted := providerscommon.CompactActions(arr1, arr2)
	assert.NotNil(t, compacted)
	assert.Equal(t, []policysupport.ActionInfo{
		{ActionUri: "1one"}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}, compacted)
}

func TestCompactActions_UniqueAndWhitespace(t *testing.T) {
	arr1 := []policysupport.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}
	arr2 := []policysupport.ActionInfo{
		{ActionUri: ""}, {ActionUri: "4four"}, {ActionUri: " "}, {ActionUri: "5five"},
	}

	compacted := providerscommon.CompactActions(arr1, arr2)
	assert.NotNil(t, compacted)
	assert.Equal(t, []policysupport.ActionInfo{
		{ActionUri: "1one"}, {ActionUri: "2two"}, {ActionUri: "3three"}, {ActionUri: "4four"}, {ActionUri: "5five"},
	}, compacted)
}

func TestCompactActions_OneEmptyNil(t *testing.T) {
	arr := []policysupport.ActionInfo{
		{ActionUri: ""}, {ActionUri: "1one"}, {ActionUri: " "}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}

	compacted := providerscommon.CompactActions(arr, nil)
	assert.NotNil(t, compacted)
	assert.Equal(t, []policysupport.ActionInfo{
		{ActionUri: "1one"}, {ActionUri: "2two"}, {ActionUri: "3three"},
	}, compacted)

	compacted = providerscommon.CompactActions(nil, arr)
	assert.NotNil(t, compacted)
	assert.Equal(t, []policysupport.ActionInfo{
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

func TestResourcePolicyMap_ReturnsEmpty(t *testing.T) {
	actMap := providerscommon.ResourcePolicyMap([]policysupport.PolicyInfo{})
	assert.NotNil(t, actMap)
	assert.Empty(t, actMap)

	actMap = providerscommon.ResourcePolicyMap(nil)
	assert.NotNil(t, actMap)
	assert.Empty(t, actMap)
}

func TestResourcePolicyMap_DupResourceDupMembers(t *testing.T) {
	pol1 := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{
			{ActionUri: ""}, {ActionUri: "1act"}, {ActionUri: " "}, {ActionUri: "2act"}},
		Subject: policysupport.SubjectInfo{Members: []string{"1mem", "", "2mem"}},
		Object:  policysupport.ObjectInfo{ResourceID: "resource1"},
	}

	pol2 := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{
			{ActionUri: ""}, {ActionUri: "3act"}, {ActionUri: " "}, {ActionUri: "4act"}},
		Subject: policysupport.SubjectInfo{Members: []string{"1mem", "", "2mem"}},
		Object:  policysupport.ObjectInfo{ResourceID: "resource1"},
	}

	orig := []policysupport.PolicyInfo{pol1, pol2}
	actMap := providerscommon.ResourcePolicyMap(orig)
	assert.NotNil(t, actMap)
	assert.Equal(t, 1, len(actMap))

	expResource := "resource1"
	expActionUris := []policysupport.ActionInfo{{ActionUri: "1act"}, {ActionUri: "2act"}, {ActionUri: "3act"}, {ActionUri: "4act"}}
	expMembers := []string{"1mem", "2mem"}

	actPol, found := actMap[expResource]
	assert.True(t, found)
	assert.NotNil(t, actPol)
	assert.Equal(t, expResource, actPol.Object.ResourceID)
	assert.Equal(t, expActionUris, actPol.Actions)
	assert.Equal(t, expMembers, actPol.Subject.Members)
}

func TestResourcePolicyMap_MergeSameResource(t *testing.T) {
	pol1a := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{
			{ActionUri: "1act"}, {ActionUri: "2act"}},
		Subject: policysupport.SubjectInfo{Members: []string{"1mem", "2mem"}},
		Object:  policysupport.ObjectInfo{ResourceID: "resource1"},
	}

	pol1b := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{
			{ActionUri: "3act"}, {ActionUri: "4act"}},
		Subject: policysupport.SubjectInfo{Members: []string{"3mem", "4mem"}},
		Object:  policysupport.ObjectInfo{ResourceID: "resource1"},
	}

	pol2 := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{
			{ActionUri: "3act"}, {ActionUri: "4act"}},
		Subject: policysupport.SubjectInfo{Members: []string{"1mem", "2mem"}},
		Object:  policysupport.ObjectInfo{ResourceID: "resource2"},
	}

	orig := []policysupport.PolicyInfo{pol1a, pol2, pol1b}
	actMap := providerscommon.ResourcePolicyMap(orig)
	assert.NotNil(t, actMap)
	assert.Equal(t, 2, len(actMap))

	expResource := "resource1"
	expActionUris := []policysupport.ActionInfo{{ActionUri: "1act"}, {ActionUri: "2act"}, {ActionUri: "3act"}, {ActionUri: "4act"}}
	expMembers := []string{"1mem", "2mem", "3mem", "4mem"}
	actPol, found := actMap[expResource]
	assert.True(t, found)
	assert.NotNil(t, actPol)
	assert.Equal(t, expResource, actPol.Object.ResourceID)
	assert.Equal(t, expActionUris, actPol.Actions)
	assert.Equal(t, expMembers, actPol.Subject.Members)

	expResource = "resource2"
	expActionUris = []policysupport.ActionInfo{{ActionUri: "3act"}, {ActionUri: "4act"}}
	expMembers = []string{"1mem", "2mem"}
	actPol, found = actMap[expResource]
	assert.True(t, found)
	assert.NotNil(t, actPol)
	assert.Equal(t, expResource, actPol.Object.ResourceID)
	assert.Equal(t, expActionUris, actPol.Actions)
	assert.Equal(t, expMembers, actPol.Subject.Members)

}
