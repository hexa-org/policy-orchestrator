package policytestsupport

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"strings"
)

const PolicyObjectResourceId = "some-resource-id"
const ActionGetProfile = "GET/profile"
const ActionGetHrUs = "GET/humanresources/us"

const UserIdGetHrUsAndProfile = "get-hr-us-profile-user-id"
const UserIdGetHrUs = "get-hr-us-user-id"
const UserIdGetProfile = "get-profile-user-id"
const UserIdUnassigned1 = "unassigned-user-id-1"
const UserIdUnassigned2 = "unassigned-user-id-2"

var UserEmailGetHrUs = MakeEmail(UserIdGetHrUs)
var UserEmailGetProfile = MakeEmail(UserIdGetProfile)
var UserEmailGetHrUsAndProfile = MakeEmail(UserIdGetHrUsAndProfile)

func MakeEmail(principalId string) string {
	emailPrefix, found := strings.CutSuffix(principalId, "-id")
	if !found {
		emailPrefix = "user-not-found"
	}
	emailPrefix = strings.ReplaceAll(emailPrefix, "-", "")
	return emailPrefix + "@stratatest.io"
}

func MakePrincipalEmailMap() map[string]string {
	return map[string]string{
		UserIdGetHrUs:           MakeEmail(UserIdGetHrUs),
		UserIdGetProfile:        MakeEmail(UserIdGetProfile),
		UserIdGetHrUsAndProfile: MakeEmail(UserIdGetHrUsAndProfile),
	}
}

type ActionMembers struct {
	MemberIds []string
	Emails    []string
}

func MakeActionMembers() map[string]ActionMembers {
	return map[string]ActionMembers{
		ActionGetHrUs: {
			MemberIds: []string{UserIdGetHrUs, UserIdGetHrUsAndProfile},
			Emails:    []string{UserEmailGetHrUs, UserEmailGetHrUsAndProfile},
		},
		ActionGetProfile: {
			MemberIds: []string{UserIdGetProfile, UserIdGetHrUsAndProfile},
			Emails:    []string{UserEmailGetProfile, UserEmailGetHrUsAndProfile},
		},
	}
}

func MakeTestPolicies(actionMembers map[string]ActionMembers) []policysupport.PolicyInfo {
	policies := make([]policysupport.PolicyInfo, 0)
	for action, members := range actionMembers {
		policies = append(policies, MakeTestPolicy(action, members))
	}
	return policies
}

func MakeTestPolicy(action string, actionMembers ActionMembers) policysupport.PolicyInfo {
	return policysupport.PolicyInfo{
		Meta:    policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{{action}},
		Subject: policysupport.SubjectInfo{Members: MakePolicyTestUsers(actionMembers)},
		Object: policysupport.ObjectInfo{
			ResourceID: PolicyObjectResourceId,
		},
	}
}

func MakePolicyTestUsers(actionMember ActionMembers) []string {
	policyUsers := make([]string, 0)
	for _, email := range actionMember.Emails {
		policyUsers = append(policyUsers, MakePolicyTestUser(email))
	}
	return policyUsers
}
func MakePolicyTestUser(emailNoPrefix string) string {
	return fmt.Sprintf("user:%s", emailNoPrefix)
}