package microsoftazure

import (
	"fmt"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
)

type AzurePolicyMapper struct {
	objectId             string
	roleIdToAppRole      map[string]azad.AzureAppRole
	existingRoleIdToAras map[string][]azad.AzureAppRoleAssignment
	azureUserEmail       map[string]string
}

func NewAzurePolicyMapper(sps azad.AzureServicePrincipals, existingAssignments []azad.AzureAppRoleAssignment, azureUserEmail map[string]string) *AzurePolicyMapper {
	if len(sps.List) == 0 {
		return &AzurePolicyMapper{}
	}

	return &AzurePolicyMapper{
		objectId:             sps.List[0].Name,
		roleIdToAppRole:      mapAppRoles(sps.List[0].AppRoles),
		existingRoleIdToAras: mapAppRoleAssignments(existingAssignments),
		azureUserEmail:       azureUserEmail}
}

func (azm *AzurePolicyMapper) ToIDQL() []hexapolicy.PolicyInfo {
	policies := make([]hexapolicy.PolicyInfo, 0)
	for appRoleId, appRole := range azm.roleIdToAppRole {
		pol := azm.appRoleAssignmentToIDQL(azm.existingRoleIdToAras[appRoleId], appRole.Value)
		policies = append(policies, pol)
	}
	return policies

}

func (azm *AzurePolicyMapper) appRoleAssignmentToIDQL(assignments []azad.AzureAppRoleAssignment, action string) hexapolicy.PolicyInfo {

	members := make([]string, 0)
	for _, oneAssignment := range assignments {
		email := azm.azureUserEmail[oneAssignment.PrincipalId]
		if email != "" {
			members = append(members, fmt.Sprintf("user:%s", email))
		}

	}

	return hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{{action}},
		Subject: hexapolicy.SubjectInfo{Members: members},
		Object:  hexapolicy.ObjectInfo{ResourceID: azm.objectId},
	}
}

func mapAppRoles(appRoles []azad.AzureAppRole) map[string]azad.AzureAppRole {
	appRolesMap := make(map[string]azad.AzureAppRole)
	for _, role := range appRoles {
		appRolesMap[role.ID] = role
	}
	return appRolesMap
}

func mapAppRoleAssignments(appRoleAssignments []azad.AzureAppRoleAssignment) map[string][]azad.AzureAppRoleAssignment {
	roleAssignmentMap := make(map[string][]azad.AzureAppRoleAssignment)
	for _, roleAssignment := range appRoleAssignments {
		roleId := roleAssignment.AppRoleId
		raArray, found := roleAssignmentMap[roleId]
		if !found {
			raArray = make([]azad.AzureAppRoleAssignment, 0)
		}

		roleAssignmentMap[roleId] = append(raArray, roleAssignment)
	}
	return roleAssignmentMap
}
