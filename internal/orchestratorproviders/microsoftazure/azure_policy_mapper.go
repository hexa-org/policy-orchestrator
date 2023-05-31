package microsoftazure

import "github.com/hexa-org/policy-orchestrator/internal/policysupport"

type AzurePolicyMapper struct {
	objectId                   string
	roleIdToAppRole            map[string]azureAppRole
	roleIdToAppRoleAssignments map[string][]AzureAppRoleAssignment
	azureUserEmail             map[string]string
}

func NewAzurePolicyMapper(sps AzureServicePrincipals, appRoleAssignments []AzureAppRoleAssignment, azureUserEmail map[string]string) *AzurePolicyMapper {
	if len(sps.List) == 0 {
		return &AzurePolicyMapper{}
	}

	return &AzurePolicyMapper{
		objectId:                   sps.List[0].ID,
		roleIdToAppRole:            mapAppRoles(sps.List[0].AppRoles),
		roleIdToAppRoleAssignments: mapAppRoleAssignments(appRoleAssignments),
		azureUserEmail:             azureUserEmail}
}

func (azm *AzurePolicyMapper) ToIDQL() []policysupport.PolicyInfo {
	policies := make([]policysupport.PolicyInfo, 0)
	for appRoleId, appRole := range azm.roleIdToAppRole {
		pol := azm.appRoleAssignmentToIDQL(azm.roleIdToAppRoleAssignments[appRoleId], appRole.Value)
		policies = append(policies, pol)
	}
	return policies

}

func (azm *AzurePolicyMapper) appRoleAssignmentToIDQL(assignments []AzureAppRoleAssignment, action string) policysupport.PolicyInfo {

	members := make([]string, 0)
	for _, oneAssignment := range assignments {
		email := azm.azureUserEmail[oneAssignment.PrincipalId]
		if email != "" {
			members = append(members, email)
		}

	}

	return policysupport.PolicyInfo{
		Meta:    policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{{action}},
		Subject: policysupport.SubjectInfo{Members: members},
		Object:  policysupport.ObjectInfo{ResourceID: azm.objectId},
	}
}

func mapAppRoles(appRoles []azureAppRole) map[string]azureAppRole {
	appRolesMap := make(map[string]azureAppRole)
	for _, role := range appRoles {
		appRolesMap[role.ID] = role
	}
	return appRolesMap
}

func mapAppRoleAssignments(appRoleAssignments []AzureAppRoleAssignment) map[string][]AzureAppRoleAssignment {
	roleAssignmentMap := make(map[string][]AzureAppRoleAssignment)
	for _, roleAssignment := range appRoleAssignments {
		roleId := roleAssignment.AppRoleId
		raArray, found := roleAssignmentMap[roleId]
		if !found {
			raArray = make([]AzureAppRoleAssignment, 0)
		}

		roleAssignmentMap[roleId] = append(raArray, roleAssignment)
	}
	return roleAssignmentMap
}
