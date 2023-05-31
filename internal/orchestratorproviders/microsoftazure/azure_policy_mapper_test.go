package microsoftazure_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAzurePolicyMapper_ToIDQL(t *testing.T) {
	principalEmails := policytestsupport.MakePrincipalEmailMap()
	roleAssignments := azuretestsupport.AppRoleAssignments
	sps := azuretestsupport.AzureServicePrincipals()
	mapper := microsoftazure.NewAzurePolicyMapper(sps, roleAssignments, principalEmails)
	actPolicies := mapper.ToIDQL()
	assert.NotNil(t, actPolicies)
	assert.Equal(t, len(sps.List[0].AppRoles), len(actPolicies))

	actPolicyMap := make(map[string][]string)
	for _, pol := range actPolicies {
		assert.Equal(t, 1, len(pol.Actions))
		actPolicyMap[pol.Actions[0].ActionUri] = pol.Subject.Members
	}

	for _, expAction := range []string{policytestsupport.ActionGetHrUs, policytestsupport.ActionGetProfile} {

		assert.NotNil(t, actPolicyMap[expAction])
		var mainEmail string
		switch expAction {
		case policytestsupport.ActionGetHrUs:
			mainEmail = policytestsupport.UserEmailGetHrUs
			break
		case policytestsupport.ActionGetProfile:
			mainEmail = policytestsupport.UserEmailGetProfile
		}
		assert.Contains(t, actPolicyMap[expAction], mainEmail)
		assert.Contains(t, actPolicyMap[expAction], policytestsupport.UserEmailGetHrUsAndProfile)
	}
}

func TestAzurePolicyMapper_ToIDQL_NoRoleAssignments(t *testing.T) {
	sps := azuretestsupport.AzureServicePrincipals()
	mapper := microsoftazure.NewAzurePolicyMapper(sps, nil, nil)
	actPolicies := mapper.ToIDQL()
	assert.NotNil(t, actPolicies)
	assert.Equal(t, len(sps.List[0].AppRoles), len(actPolicies))

	actPolicyMap := make(map[string][]string)
	for _, pol := range actPolicies {
		assert.Equal(t, 1, len(pol.Actions))
		actPolicyMap[pol.Actions[0].ActionUri] = pol.Subject.Members
	}

	for _, expAction := range []string{policytestsupport.ActionGetHrUs, policytestsupport.ActionGetProfile} {
		assert.NotNil(t, actPolicyMap[expAction])
		assert.Equal(t, 0, len(actPolicyMap[expAction]))
	}
}

func TestAzurePolicyMapper_ToIDQL_NoAppRoles(t *testing.T) {
	mapper := microsoftazure.NewAzurePolicyMapper(microsoftazure.AzureServicePrincipals{}, nil, nil)
	actPolicies := mapper.ToIDQL()
	assert.NotNil(t, actPolicies)
	assert.Equal(t, 0, len(actPolicies))
}
