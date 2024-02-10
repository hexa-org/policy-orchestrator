package policytestsupport

import (
	"fmt"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"strings"
)

const PolicyObjectResourceId = "some-resource-id"
const ResourceHrUs = "/humanresources/us"
const ResourceProfile = "/profile"
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

func MakeTestPolicies(actionMembers map[string]ActionMembers) []hexapolicy.PolicyInfo {
	policies := make([]hexapolicy.PolicyInfo, 0)
	for action, members := range actionMembers {
		policies = append(policies, MakeTestPolicy(PolicyObjectResourceId, action, members))
	}
	return policies
}

// MakeRoleSubjectTestPolicies - makes policies with passed in param
// actionMembers = { "GET/humanresources/us": ["role1", "role2"] }
func MakeRoleSubjectTestPolicies(actionMembers map[string][]string) []hexapolicy.PolicyInfo {
	policies := make([]hexapolicy.PolicyInfo, 0)
	for action, members := range actionMembers {
		parts := strings.Split(action, "/")
		actionUri := "http:" + parts[0]
		resId := "/" + strings.Join(parts[1:], "/")
		policies = append(policies, MakeRoleSubjectTestPolicy(resId, actionUri, members))
	}
	return policies
}

func MakeRoleSubjectTestPolicy(resourceId string, action string, roles []string) hexapolicy.PolicyInfo {
	return hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{{action}},
		Subject: hexapolicy.SubjectInfo{Members: roles},
		Object: hexapolicy.ObjectInfo{
			ResourceID: resourceId,
		},
	}
}

func MakeTestPolicy(resourceId string, action string, actionMembers ActionMembers) hexapolicy.PolicyInfo {
	return hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: []hexapolicy.ActionInfo{{action}},
		Subject: hexapolicy.SubjectInfo{Members: MakePolicyTestUsers(actionMembers)},
		Object: hexapolicy.ObjectInfo{
			ResourceID: resourceId,
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
