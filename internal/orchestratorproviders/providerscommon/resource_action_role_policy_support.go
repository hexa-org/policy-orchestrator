package providerscommon

import (
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/functionalsupport"
	log "golang.org/x/exp/slog"
)

// CalcResourceActionRolesForUpdate
// Builds ResourceActionRoles that need to be updated only for those policies that match an existing resource action.
// If existing is empty, returns empty slice
// If policyInfos is empty, returns empty slice
func CalcResourceActionRolesForUpdate(existing []ResourceActionRoles, policyInfos []policysupport.PolicyInfo) []ResourceActionRoles {

	existingRarMap := mapResourceActionRoles(existing)
	newPolicies := ResourcePolicyMap(policyInfos)

	if len(existingRarMap) == 0 || len(newPolicies) == 0 {
		return []ResourceActionRoles{}
	}

	rarUpdateList := make([]ResourceActionRoles, 0)

	for polResource, pol := range newPolicies {
		polAction := pol.Actions[0].ActionUri
		polRoles := pol.Subject.Members

		newRarKey := MakeRarKeyForPolicy(polAction, polResource)
		existingRar, found := existingRarMap[newRarKey]
		if !found {
			log.Warn("Ignoring policy as no existing resource action matches", "resource", polResource, "action", polAction)
			continue
		}

		keepRoles := rolesToKeep(existingRar.Roles, polRoles)
		nKeep := len(keepRoles)
		if nKeep == len(existingRar.Roles) && nKeep == len(polRoles) {
			// no changes
			continue
		}

		updatedRar := NewResourceActionRoles(polResource, polAction, keepRoles)
		rarUpdateList = append(rarUpdateList, updatedRar)
	}

	return rarUpdateList
}

// rolesToKeep
// removes from existingRoles, those that are not present in newRoles
// returns union of roles
// 1) that are present in both
// 2) newRoles - existingRoles
func rolesToKeep(existingRoles, newRoles []string) []string {
	_, keepRoles, newOnly := functionalsupport.DiffUnique(existingRoles, newRoles)
	keep := make([]string, 0)
	keep = append(keep, keepRoles...)          // keep matching
	keep = append(keep, newOnly...)            // keep new ones
	keep = functionalsupport.SortCompact(keep) // keep sorted and compact
	return keep
}

func mapResourceActionRoles(rarList []ResourceActionRoles) map[string]ResourceActionRoles {
	rarMap := make(map[string]ResourceActionRoles)
	for _, rar := range rarList {
		rarMap[rar.Name()] = rar
	}
	return rarMap
}
