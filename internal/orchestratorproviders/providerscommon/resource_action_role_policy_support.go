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

		hasChanges, rolesToUpdate := findRolesToUpdate(existingRar.Roles, polRoles)
		if !hasChanges {
			continue
		}

		updatedRar := NewResourceActionUriRoles(polResource, polAction, rolesToUpdate)
		rarUpdateList = append(rarUpdateList, updatedRar)
	}

	return rarUpdateList
}

// rolesToKeep - removes from existingRoles, those that are not present in newRoles
// Returns bool, slice
// false indicates no changes (i.e. when existing == new)
// slice - list of changes (i.e. new - existing)
// OR nil/empty if all existing need to be removed.
func findRolesToUpdate(existingRoles, newRoles []string) (hasChanges bool, changes []string) {
	existingOnly, matches, newOnly := functionalsupport.DiffUnique(existingRoles, newRoles)
	if len(existingOnly) == 0 && len(newOnly) == 0 {
		// no changes
		return
	}

	hasChanges = true
	changes = append(changes, matches...)            // keep matching
	changes = append(changes, newOnly...)            // keep new ones
	changes = functionalsupport.SortCompact(changes) // keep sorted and compact
	return
}

func mapResourceActionRoles(rarList []ResourceActionRoles) map[string]ResourceActionRoles {
	rarMap := make(map[string]ResourceActionRoles)
	for _, rar := range rarList {
		rarMap[rar.Name()] = rar
	}
	return rarMap
}
