package policytestsupport

import "strings"

const ProtectedApiResourceId = "some-resource-id"
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
