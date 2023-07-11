package providerscommon

import (
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/functionalsupport"
)

const ActionUriPrefix = "http:"

func BuildPolicies(resourceActionRolesList []ResourceActionRoles) []policysupport.PolicyInfo {
	policies := make([]policysupport.PolicyInfo, 0)
	for _, one := range resourceActionRolesList {
		httpMethod := one.Action
		policies = append(policies, policysupport.PolicyInfo{
			Meta:    policysupport.MetaInfo{Version: "0.5"},
			Actions: []policysupport.ActionInfo{{ActionUriPrefix + httpMethod}},
			Subject: policysupport.SubjectInfo{Members: one.Roles},
			Object:  policysupport.ObjectInfo{ResourceID: one.Resource},
		})
	}
	return policies
}

// ResourcePolicyMap - makes a map of resource -> PolicyInfo
// If multiple PolicyInfo elements exist for a given resource, these are merged
// This ensures downstream functions do not have to deal with multiple policies for same resource.
// Also filters out any empty strings or duplicates in members or actions
func ResourcePolicyMap(origPolicies []policysupport.PolicyInfo) map[string]policysupport.PolicyInfo {
	resPolicyMap := make(map[string]policysupport.PolicyInfo)
	for _, pol := range origPolicies {
		resource := pol.Object.ResourceID

		var existingActions []policysupport.ActionInfo
		var existingMembers []string
		if existing, exists := resPolicyMap[resource]; exists {
			existingActions = existing.Actions
			existingMembers = existing.Subject.Members
		}

		mergedActions := CompactActions(existingActions, pol.Actions)
		newMembers := CompactMembers(existingMembers, pol.Subject.Members)

		newPol := policysupport.PolicyInfo{
			Meta:    policysupport.MetaInfo{Version: "0.5"},
			Actions: mergedActions,
			Subject: policysupport.SubjectInfo{Members: newMembers},
			Object:  policysupport.ObjectInfo{ResourceID: resource},
		}

		resPolicyMap[resource] = newPol

	}
	return resPolicyMap
}

func CompactActions(existing, new []policysupport.ActionInfo) []policysupport.ActionInfo {
	actionUris := make([]string, 0)
	for _, act := range existing {
		actionUris = append(actionUris, act.ActionUri)
	}
	for _, act := range new {
		actionUris = append(actionUris, act.ActionUri)
	}
	actionUris = functionalsupport.SortCompact(actionUris)

	actionInfos := make([]policysupport.ActionInfo, 0)
	for _, uri := range actionUris {
		actionInfos = append(actionInfos, policysupport.ActionInfo{
			ActionUri: uri,
		})
	}
	return actionInfos
}

func CompactMembers(existing, new []string) []string {
	compacted := make([]string, 0)
	compacted = append(compacted, existing...)
	compacted = append(compacted, new...)
	return functionalsupport.SortCompact(compacted)
}
