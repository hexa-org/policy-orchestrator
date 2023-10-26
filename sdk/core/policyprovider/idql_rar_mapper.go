package policyprovider

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	log "golang.org/x/exp/slog"
	"strings"
)

// mapIdqlToRar - converts IDQL policy to a map with
// key = resource+action, value = rar
// each action in a IDQL results in a new rar.
func mapIdqlToRar(origPolicies ...hexapolicy.PolicyInfo) (map[string]rar.ResourceActionRoles, error) {
	resActionRarMap := make(map[string]rar.ResourceActionRoles)
	for _, pol := range origPolicies {
		resource := strings.TrimSpace(pol.Object.ResourceID)
		if resource == "" {
			return nil, errors.New("mapIdqlToRar error mapping IDQL with empty resource")
		}

		if len(pol.Actions) == 0 {
			return nil, fmt.Errorf("mapIdqlToRar error mapping IDQL with nil actionUri. Resource=%s", resource)
		}

		// convert each action to a rar
		for _, anAction := range pol.Actions {
			actionUri := strings.TrimSpace(anAction.ActionUri)
			if actionUri == "" {
				return nil, fmt.Errorf("mapIdqlToRar error mapping IDQL without actionUri. Resource=%s", resource)
			}

			lookupKey := actionUri + resource
			matchingRar, _ := resActionRarMap[lookupKey]
			members := make([]string, 0)
			members = append(members, matchingRar.Members()...)
			members = append(members, pol.Subject.Members...)
			newRar, nErr := rar.NewResourceActionUriRoles(resource, []string{actionUri}, members)
			if nErr != nil {
				log.Error("mapIdqlToRar",
					"failed to make ResourceActionRoles resource", resource,
					"action", actionUri,
					"members", members,
					"error", nErr)
				return nil, nErr
			}
			resActionRarMap[lookupKey] = newRar
		}
	}

	return resActionRarMap, nil
}
