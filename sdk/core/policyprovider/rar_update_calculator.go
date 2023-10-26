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
		newRar, found := c.newRarMap[lookupKey]

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
		updateList = append(updateList, newRar)
	}

	return updateList
}
