package policytestsupport

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"sort"
	"strings"
)

func MakeRarList(retActionRoles map[string][]string) []providerscommon.ResourceActionRoles {
	rarList := make([]providerscommon.ResourceActionRoles, 0)

	for actionAndRes, roles := range retActionRoles {
		resRole := MakeRar(actionAndRes, roles)
		rarList = append(rarList, resRole)
	}

	sort.SliceStable(rarList, func(i, j int) bool {
		a := rarList[i]
		b := rarList[j]
		resComp := strings.Compare(a.Resource, b.Resource)
		actComp := strings.Compare(a.Action, b.Action)
		switch resComp {
		case 0:
			return actComp <= 0
		default:
			return resComp < 0
		}
	})

	/*slices.SortStableFunc(rarList, func(a, b providerscommon.ResourceActionRoles) bool {
		resComp := strings.Compare(a.Resource, b.Resource)
		actComp := strings.Compare(a.Action, b.Action)
		switch resComp {
		case 0:
			return actComp <= 0
		default:
			return resComp < 0
		}
	})*/
	return rarList
}

func MakeRar(actionAndRes string, roles []string) providerscommon.ResourceActionRoles {
	parts := strings.Split(actionAndRes, "/")
	resActionKey := fmt.Sprintf("resrol-http%s-%s", strings.ToLower(parts[0]), strings.Join(parts[1:], "-"))
	return providerscommon.NewResourceActionRolesFromProviderValue(resActionKey, roles)
}
