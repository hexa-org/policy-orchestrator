package policyprovider

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var expResource1 = testhelper.ResourceHrUs
var expMethods1 = []string{http.MethodGet, http.MethodPost}
var existingMembers1 = []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}

var expResource2 = testhelper.ResourceProfile
var expMethods2 = []string{http.MethodPut, http.MethodDelete}
var existingMembers2 = []string{testhelper.RoleReadHrUs, testhelper.RoleReadProfile}

// TestCalculate_AddMembers asserts update calculator returns expected rars
// when adding new members to existing
func TestCalculate_AddMembers(t *testing.T) {

	// 4 rars with same members
	existingRars := testhelper.MakeRarListMultiple(
		[]string{expResource1, expResource2},
		[][]string{expMethods1, expMethods2},
		[][]string{existingMembers1, existingMembers2})

	// Add 2 new members to resource1
	newMembers1 := addMembersToExisting(existingMembers1, testhelper.RoleUnassigned1, testhelper.RoleUnassigned2)
	// Add 2 new members to resource2
	newMembers2 := addMembersToExisting(existingMembers2, testhelper.RoleUnassigned2, testhelper.RoleUnassigned1)

	// 4 rars with new members (all same)
	newRarMap := testhelper.MakeRarMapMultiple([]string{expResource1, expResource2}, [][]string{expMethods1, expMethods2}, [][]string{newMembers1, newMembers2})

	calc := newUpdateCalculator(existingRars, newRarMap)
	updateList := calc.calculate()
	assert.Len(t, updateList, len(existingRars))

	// Expect 4 rars with new and existing members
	expUpdateList := testhelper.MakeRarListMultiple(
		[]string{expResource1, expResource2},
		[][]string{expMethods1, expMethods2},
		[][]string{newMembers1, newMembers2})

	assert.Equal(t, expUpdateList, updateList)

}

// TestCalculate_AddMembers asserts update calculator returns expected rars
// when replacing all members
func TestCalculate_RemoveMembers(t *testing.T) {

	tests := []struct {
		name        string
		existing1   []string
		existing2   []string
		newMembers1 []string
		newMembers2 []string
	}{
		{
			name:        "Remove Some Members",
			existing1:   addMembersToExisting(existingMembers1, testhelper.RoleUnassigned1, testhelper.RoleUnassigned2),
			existing2:   addMembersToExisting(existingMembers2, testhelper.RoleUnassigned2, testhelper.RoleUnassigned1),
			newMembers1: existingMembers1,
			newMembers2: existingMembers2,
		},
		{
			name:        "Replace All Members",
			existing1:   existingMembers1,
			existing2:   existingMembers2,
			newMembers1: []string{testhelper.RoleUnassigned1, testhelper.RoleUnassigned2},
			newMembers2: []string{testhelper.RoleUnassigned2, testhelper.RoleUnassigned1},
		},
		{
			name:        "Remove All Members",
			existing1:   existingMembers1,
			existing2:   existingMembers2,
			newMembers1: nil,
			newMembers2: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 4 rars with same 4 members each
			existingRars := testhelper.MakeRarListMultiple(
				[]string{expResource1, expResource2},
				[][]string{expMethods1, expMethods2},
				[][]string{tt.existing1, tt.existing2})

			// 4 rars with all members replaced
			newRarMap := testhelper.MakeRarMapMultiple([]string{expResource1, expResource2}, [][]string{expMethods1, expMethods2}, [][]string{tt.newMembers1, tt.newMembers2})
			calc := newUpdateCalculator(existingRars, newRarMap)
			updateList := calc.calculate()
			assert.Len(t, updateList, len(existingRars))

			// Expect 4 rars with new members only
			expUpdateList := testhelper.MakeRarListMultiple(
				[]string{expResource1, expResource2},
				[][]string{expMethods1, expMethods2},
				[][]string{tt.newMembers1, tt.newMembers2})

			assert.Equal(t, expUpdateList, updateList)
		})
	}
}

// TestCalculate_SkipOnNoChanges asserts no updates are returned
// when there are no changes between exising and new
func TestCalculate_SkipOnNoChanges(t *testing.T) {
	// 4 rars with same members
	existingRars := testhelper.MakeRarListMultiple(
		[]string{expResource1, expResource2},
		[][]string{expMethods1, expMethods2},
		[][]string{existingMembers1, existingMembers2})

	tests := []struct {
		name               string
		newMembers1        []string
		newMembers2        []string
		expEmptyUpdateList bool
	}{
		{
			name:               "one changed, one unchanged",
			newMembers1:        addMembersToExisting(existingMembers1, testhelper.RoleUnassigned1, testhelper.RoleUnassigned2),
			newMembers2:        existingMembers2,
			expEmptyUpdateList: false,
		},
		{
			name:               "no changes",
			newMembers1:        existingMembers1,
			newMembers2:        existingMembers2,
			expEmptyUpdateList: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 4 rars with new members (all same)
			newRarMap := testhelper.MakeRarMapMultiple([]string{expResource1, expResource2}, [][]string{expMethods1, expMethods2}, [][]string{tt.newMembers1, tt.newMembers2})

			calc := newUpdateCalculator(existingRars, newRarMap)
			updateList := calc.calculate()
			if tt.expEmptyUpdateList {
				assert.Empty(t, updateList)
			} else {
				assert.Len(t, updateList, 2)
				// Expect 4 rars with new members only
				expUpdateList := testhelper.MakeRarListMultiple(
					[]string{expResource1},
					[][]string{expMethods1},
					[][]string{tt.newMembers1})

				assert.Equal(t, expUpdateList, updateList)
			}
		})
	}
}

func addMembersToExisting(existing []string, newMembers ...string) []string {
	newMembers1 := make([]string, 0)
	newMembers1 = append(newMembers1, existing...)
	newMembers1 = append(newMembers1, newMembers...)
	return newMembers
}
