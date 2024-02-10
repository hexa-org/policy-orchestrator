package policyprovider

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"golang.org/x/exp/slices"
	log "golang.org/x/exp/slog"
)

type updateCalculator struct {
	existingRars []rar.ResourceActionRoles
	newRarMap    map[string]rar.ResourceActionRoles
}

func newUpdateCalculator(existingRars []rar.ResourceActionRoles, newRarMap map[string]rar.ResourceActionRoles) updateCalculator {
	return updateCalculator{existingRars: existingRars, newRarMap: newRarMap}
}

func (c updateCalculator) calculate() []rar.ResourceActionRoles {
	log.Info("updateCalculator.calculate", "msg", "BEGIN")

	// finds new - existing
	// if no changes, do not add to update list
	// if new is empty, use new members
	// if existing is empty, new members
	// if both are non-empty,
	//   no change - do no add to update list
	//   there are changes - use new members.
	// if both are empty, no change

	updateList := make([]rar.ResourceActionRoles, 0)
	for _, existing := range c.existingRars {

		lookupKey := existing.Actions()[0] + existing.Resource() // TODO handle array actions
		log.Info("updateCalculator.calculate", "msg", "LOOP existing lookupKey", lookupKey)

		newRar, found := c.newRarMap[lookupKey]
		log.Info("updateCalculator.calculate", "msg", "LOOP newRar", newRar, "found", found)
		// We DON'T support orchestrating policy fragments
		if !found {
			log.Warn("updateCalculator.calculate", "msg", "requested policies do not contain existing policy",
				"resource", existing.Resource(),
				"action", existing.Actions(),
				"members", existing.Members())

			continue
		}

		newMembersLen := len(newRar.Members())
		existingMembersLen := len(existing.Members())
		log.Info("updateCalculator.calculate", "newMembersLen", newMembersLen, "existingMembersLen", existingMembersLen)

		if newMembersLen == existingMembersLen {
			// If both empty OR both same
			if newMembersLen == 0 || slices.Compare(existing.Members(), newRar.Members()) == 0 {
				log.Info("updateCalculator.calculate", "skipping", "no changes",
					"resource", existing.Resource(),
					"action", existing.Actions(),
					"members", existing.Members())
				continue // no change
			}
		}

		// if new members empty OR if existing members are empty, OR if both are non-empty but differ
		// we should just be able to use the new
		log.Info("updateCalculator.calculate", "msg", "add to update list")
		updateList = append(updateList, newRar)
	}

	return updateList
}
