package policyprovider

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	log "golang.org/x/exp/slog"
	"strings"
)

const ActionUriPrefix = "http:"

// mapIdqlToRar - converts IDQL policy to a map with
// key = resource+action, value = rar
// each action in a IDQL results in a new rar.
func mapIdqlToRar(origPolicies ...hexapolicy.PolicyInfo) (map[string]rar.ResourceActionRoles, error) {
	resActionRarMap := make(map[string]rar.ResourceActionRoles)
	log.Info("policyprovider.mapIdqlToRar", "origPolicies", origPolicies)

	for _, pol := range origPolicies {

		resource := strings.TrimSpace(pol.Object.ResourceID)
		log.Info("policyprovider.mapIdqlToRar", "LOOP onePol.resource", resource, "actions", pol.Actions)
		if resource == "" {
			return nil, errors.New("mapIdqlToRar error mapping IDQL with empty resource")
		}

		if len(pol.Actions) == 0 {
			return nil, fmt.Errorf("mapIdqlToRar error mapping IDQL with nil actionUri. Resource=%s", resource)
		}

		// convert each action to a rar
		for _, anAction := range pol.Actions {
			log.Info("policyprovider.mapIdqlToRar", "LOOP actions", anAction.ActionUri)
			actionUri := strings.TrimSpace(anAction.ActionUri)
			actionUri = strings.TrimPrefix(actionUri, ActionUriPrefix)
			if actionUri == "" {
				return nil, fmt.Errorf("mapIdqlToRar error mapping IDQL without actionUri. Resource=%s", resource)
			}

			lookupKey := actionUri + resource
			log.Info("policyprovider.mapIdqlToRar", "LOOP actions lookupKey", lookupKey)
			matchingRar, _ := resActionRarMap[lookupKey]
			log.Info("policyprovider.mapIdqlToRar", "LOOP matchingRar", matchingRar)
			members := make([]string, 0)
			members = append(members, matchingRar.Members()...)
			members = append(members, pol.Subject.Members...)
			log.Info("policyprovider.mapIdqlToRar", "newMembers", members)
			newRar, nErr := rar.NewResourceActionUriRoles(resource, []string{actionUri}, members)
			log.Info("policyprovider.mapIdqlToRar", "newRar", newRar)
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
