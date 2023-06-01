package azuretestsupport

import (
	"github.com/google/uuid"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"sort"
)

var AppRoleAssignmentGetHrUsAndProfile = []microsoftazure.AzureAppRoleAssignment{
	NewAppRoleAssignments(AppRoleIdGetHrUs, policytestsupport.UserIdGetHrUsAndProfile),
	NewAppRoleAssignments(AppRoleIdGetProfile, policytestsupport.UserIdGetHrUsAndProfile),
}

var AppRoleAssignmentGetHrUs = []microsoftazure.AzureAppRoleAssignment{
	NewAppRoleAssignments(AppRoleIdGetHrUs, policytestsupport.UserIdGetHrUs),
}

var AppRoleAssignmentGetProfile = []microsoftazure.AzureAppRoleAssignment{
	NewAppRoleAssignments(AppRoleIdGetProfile, policytestsupport.UserIdGetProfile),
}

var AppRoleAssignmentMultipleMembers = []microsoftazure.AzureAppRoleAssignment{
	NewAppRoleAssignments(AppRoleIdGetHrUs, policytestsupport.UserIdGetHrUs),
	NewAppRoleAssignments(AppRoleIdGetHrUs, policytestsupport.UserIdGetHrUsAndProfile),
}

var AppRoleAssignmentForAdd = []microsoftazure.AzureAppRoleAssignment{
	NewAppRoleAssignments(AppRoleIdGetProfile, policytestsupport.UserIdUnassigned1),
	NewAppRoleAssignments(AppRoleIdGetProfile, policytestsupport.UserIdUnassigned2),
}

var AppRoleAssignments = []microsoftazure.AzureAppRoleAssignment{
	NewAppRoleAssignments(AppRoleIdGetHrUs, policytestsupport.UserIdGetHrUs),
	NewAppRoleAssignments(AppRoleIdGetProfile, policytestsupport.UserIdGetProfile),
	NewAppRoleAssignments(AppRoleIdGetHrUs, policytestsupport.UserIdGetHrUsAndProfile),
	NewAppRoleAssignments(AppRoleIdGetProfile, policytestsupport.UserIdGetHrUsAndProfile),
}

func NewAppRoleAssignments(appRoleId AppRoleId, principalId string) microsoftazure.AzureAppRoleAssignment {
	return microsoftazure.AzureAppRoleAssignment{
		ID:          uuid.NewString(),
		AppRoleId:   string(appRoleId),
		PrincipalId: principalId,
		ResourceId:  ServicePrincipalId,
	}
}

func MakeAssignments(assignments []microsoftazure.AzureAppRoleAssignment) microsoftazure.AzureAppRoleAssignments {
	return microsoftazure.AzureAppRoleAssignments{List: assignments}
}

func AssignmentsWithoutId(assignments []microsoftazure.AzureAppRoleAssignment) []microsoftazure.AzureAppRoleAssignment {
	newAssignments := make([]microsoftazure.AzureAppRoleAssignment, 0)
	for _, ara := range assignments {
		newAra := microsoftazure.AzureAppRoleAssignment{
			AppRoleId:   ara.AppRoleId,
			PrincipalId: ara.PrincipalId,
			ResourceId:  ara.ResourceId,
		}

		newAssignments = append(newAssignments, newAra)
	}
	return newAssignments
}

func AssignmentsForDelete(assignments []microsoftazure.AzureAppRoleAssignment) []microsoftazure.AzureAppRoleAssignment {
	newAssignments := make([]microsoftazure.AzureAppRoleAssignment, 0)
	for _, ara := range assignments {
		newAra := microsoftazure.AzureAppRoleAssignment{
			AppRoleId:  ara.AppRoleId,
			ResourceId: ara.ResourceId,
		}

		newAssignments = append(newAssignments, newAra)
	}
	return newAssignments
}

func MakePolicies(assignments []microsoftazure.AzureAppRoleAssignment) []policysupport.PolicyInfo {
	policyMapper := microsoftazure.NewAzurePolicyMapper(AzureServicePrincipals(),
		assignments,
		policytestsupport.MakePrincipalEmailMap())

	return policyMapper.ToIDQL()
}

func SortAssignments(orig []microsoftazure.AzureAppRoleAssignment) []microsoftazure.AzureAppRoleAssignment {
	sorted := make([]microsoftazure.AzureAppRoleAssignment, 0)
	sorted = append(sorted, orig...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].AppRoleId == sorted[j].AppRoleId {
			return sorted[i].PrincipalId <= sorted[j].PrincipalId
		}

		return sorted[i].AppRoleId < sorted[j].AppRoleId
	})
	return sorted
}
