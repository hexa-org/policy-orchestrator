package providerscommon

import (
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/functionalsupport"
	"golang.org/x/exp/slices"
	log "golang.org/x/exp/slog"
	"sort"
	"strings"
)

const ActionUriPrefix = "http:"

func BuildPolicies(resourceActionRolesList []ResourceActionRoles) []policysupport.PolicyInfo {
	policies := make([]policysupport.PolicyInfo, 0)
	for _, one := range resourceActionRolesList {
		httpMethod := one.Action
		roles := one.Roles
		slices.Sort(roles)
		policies = append(policies, policysupport.PolicyInfo{
			Meta:    policysupport.MetaInfo{Version: "0.5"},
			Actions: []policysupport.ActionInfo{{ActionUriPrefix + httpMethod}},
			Subject: policysupport.SubjectInfo{Members: roles},
			Object:  policysupport.ObjectInfo{ResourceID: one.Resource},
		})
	}

	sortPolicies(policies)
	return policies
}

func FlattenPolicy(origPolicies []policysupport.PolicyInfo) []policysupport.PolicyInfo {

	resActionPolicyMap := make(map[string]policysupport.PolicyInfo)
	for _, pol := range origPolicies {
		resource := pol.Object.ResourceID
		if resource == "" {
			log.Warn("FlattenPolicy Skipping policy without resource")
			continue
		}
		for _, act := range pol.Actions {
			if strings.TrimSpace(act.ActionUri) == "" {
				log.Warn("FlattenPolicy Skipping policy without actionUri", "resource", resource)
				continue
			}
			lookupKey := act.ActionUri + resource
			matchingPolicy, found := resActionPolicyMap[lookupKey]
			var existingMembers []string
			if found {
				existingMembers = matchingPolicy.Subject.Members
			}
			newMembers := CompactMembers(existingMembers, pol.Subject.Members)
			newPol := policysupport.PolicyInfo{
				Meta:    policysupport.MetaInfo{Version: "0.5"},
				Actions: []policysupport.ActionInfo{{ActionUri: act.ActionUri}},
				Subject: policysupport.SubjectInfo{Members: newMembers},
				Object:  policysupport.ObjectInfo{ResourceID: resource},
			}

			resActionPolicyMap[lookupKey] = newPol
		}
	}

	flat := make([]policysupport.PolicyInfo, 0)
	for _, pol := range resActionPolicyMap {
		flat = append(flat, pol)
	}

	sortPolicies(flat)
	return flat
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

func sortPolicies(policies []policysupport.PolicyInfo) {
	sort.SliceStable(policies, func(i, j int) bool {
		resComp := strings.Compare(policies[i].Object.ResourceID, policies[j].Object.ResourceID)
		actComp := strings.Compare(policies[i].Actions[0].ActionUri, policies[j].Actions[0].ActionUri)
		switch resComp {
		case 0:
			return actComp <= 0
		default:
			return resComp < 0

		}
	})
}

/*
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
*/
