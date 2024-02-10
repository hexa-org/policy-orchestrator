package providerscommon

import (
	"sort"
	"strings"

	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/functionalsupport"
	"golang.org/x/exp/slices"
	log "golang.org/x/exp/slog"
)

const ActionUriPrefix = "http:"

func BuildPolicies(resourceActionRolesList []ResourceActionRoles) []hexapolicy.PolicyInfo {
	policies := make([]hexapolicy.PolicyInfo, 0)
	for _, one := range resourceActionRolesList {
		httpMethod := one.Action
		roles := one.Roles
		slices.Sort(roles)
		policies = append(policies, hexapolicy.PolicyInfo{
			Meta:    hexapolicy.MetaInfo{Version: "0.5"},
			Actions: []hexapolicy.ActionInfo{{ActionUriPrefix + httpMethod}},
			Subject: hexapolicy.SubjectInfo{Members: roles},
			Object:  hexapolicy.ObjectInfo{ResourceID: one.Resource},
		})
	}

	sortPolicies(policies)
	return policies
}

func FlattenPolicy(origPolicies []hexapolicy.PolicyInfo) []hexapolicy.PolicyInfo {

	resActionPolicyMap := make(map[string]hexapolicy.PolicyInfo)
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
			newPol := hexapolicy.PolicyInfo{
				Meta:    hexapolicy.MetaInfo{Version: "0.5"},
				Actions: []hexapolicy.ActionInfo{{ActionUri: act.ActionUri}},
				Subject: hexapolicy.SubjectInfo{Members: newMembers},
				Object:  hexapolicy.ObjectInfo{ResourceID: resource},
			}

			resActionPolicyMap[lookupKey] = newPol
		}
	}

	flat := make([]hexapolicy.PolicyInfo, 0)
	for _, pol := range resActionPolicyMap {
		flat = append(flat, pol)
	}

	sortPolicies(flat)
	return flat
}

func CompactActions(existing, new []hexapolicy.ActionInfo) []hexapolicy.ActionInfo {
	actionUris := make([]string, 0)
	for _, act := range existing {
		actionUris = append(actionUris, act.ActionUri)
	}
	for _, act := range new {
		actionUris = append(actionUris, act.ActionUri)
	}
	actionUris = functionalsupport.SortCompact(actionUris)

	actionInfos := make([]hexapolicy.ActionInfo, 0)
	for _, uri := range actionUris {
		actionInfos = append(actionInfos, hexapolicy.ActionInfo{
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

func sortPolicies(policies []hexapolicy.PolicyInfo) {
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
func ResourcePolicyMap(origPolicies []hexapolicy.PolicyInfo) map[string]hexapolicy.PolicyInfo {
	resPolicyMap := make(map[string]hexapolicy.PolicyInfo)
	for _, pol := range origPolicies {
		resource := pol.Object.ResourceID

		var existingActions []hexapolicy.ActionInfo
		var existingMembers []string
		if existing, exists := resPolicyMap[resource]; exists {
			existingActions = existing.Actions
			existingMembers = existing.Subject.Members
		}

		mergedActions := CompactActions(existingActions, pol.Actions)
		newMembers := CompactMembers(existingMembers, pol.Subject.Members)

		newPol := hexapolicy.PolicyInfo{
			Meta:    hexapolicy.MetaInfo{Version: "0.5"},
			Actions: mergedActions,
			Subject: hexapolicy.SubjectInfo{Members: newMembers},
			Object:  hexapolicy.ObjectInfo{ResourceID: resource},
		}

		resPolicyMap[resource] = newPol

	}
	return resPolicyMap
}
*/
