package policytestsupport

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sort"
	"testing"
)

type PolicyChecker struct {
	policies []policysupport.PolicyInfo
}

func ContainsPolicies(t *testing.T, expPolicies []policysupport.PolicyInfo, actPolicies []policysupport.PolicyInfo) bool {
	for _, act := range actPolicies {
		if HasPolicy(expPolicies, act) {
			return true
		}
	}

	return assert.Fail(t, fmt.Sprintf("Policies do not match expected: \n expected: %s\n actual: %s", expPolicies, actPolicies))
}

func HasPolicy(expPolicies []policysupport.PolicyInfo, act policysupport.PolicyInfo) bool {
	for _, exp := range expPolicies {
		if MatchPolicy(exp, act) {
			return true
		}
	}
	return false
}

func MatchPolicy(exp policysupport.PolicyInfo, act policysupport.PolicyInfo) bool {
	if exp.Object.ResourceID != act.Object.ResourceID {
		return false
	}

	expActions := sortAction(exp.Actions)
	actActions := sortAction(act.Actions)
	if !reflect.DeepEqual(expActions, actActions) {
		return false
	}

	expMembers := sortMembers(exp.Subject)
	actMembers := sortMembers(act.Subject)
	return reflect.DeepEqual(expMembers, actMembers)
}

func sortAction(orig []policysupport.ActionInfo) []policysupport.ActionInfo {
	sorted := make([]policysupport.ActionInfo, 0)
	sorted = append(sorted, orig...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ActionUri <= sorted[j].ActionUri
	})
	return sorted
}

func sortMembers(subInfo policysupport.SubjectInfo) policysupport.SubjectInfo {
	sorted := make([]string, 0)
	for _, one := range subInfo.Members {
		sorted = append(sorted, one)
	}
	sort.Strings(sorted)
	return policysupport.SubjectInfo{
		Members: sorted,
	}
}
