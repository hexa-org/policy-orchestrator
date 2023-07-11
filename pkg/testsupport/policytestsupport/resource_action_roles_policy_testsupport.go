package policytestsupport

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"strings"
)

func MakeRarList(retActionRoles map[string][]string) []providerscommon.ResourceActionRoles {
	rarList := make([]providerscommon.ResourceActionRoles, 0)

	for actionAndRes, roles := range retActionRoles {
		resRole := MakeRar(actionAndRes, roles)
		rarList = append(rarList, resRole)
	}

	return rarList
}

func MakeRar(actionAndRes string, roles []string) providerscommon.ResourceActionRoles {
	parts := strings.Split(actionAndRes, "/")
	resActionKey := fmt.Sprintf("resrol-http%s-%s", strings.ToLower(parts[0]), strings.Join(parts[1:], "-"))
	return providerscommon.NewResourceActionRolesFromProviderValue(resActionKey, roles)
}
