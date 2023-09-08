package idql

import (
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/internal/testhelper"
)

const MetaVersion = "5.0"

func ReadHrUSPolicy() hexapolicy.PolicyInfo {
	return MakeTestPolicy(testhelper.ResourceHrUs, []string{testhelper.ActionHttpGet}, []string{testhelper.RoleReadHrUs})
}

// MakeTestPolicy - creates a IDQL policy for the specified resource, http method and members
// e.g. MakeTestPolicy("some-resource",
//
//	[]string{http.MethodGet, http.MethodPost},
//	[]string{"role1", "role2"})
func MakeTestPolicy(resourceId string, httpMethods []string, members []string) hexapolicy.PolicyInfo {
	return hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: MetaVersion},
		Actions: MakeActionInfo(httpMethods...),
		Subject: hexapolicy.SubjectInfo{Members: members},
		Object: hexapolicy.ObjectInfo{
			ResourceID: resourceId,
		},
	}
}

// MakeActionInfo - converts an http method to hexapolicy.ActionInfo
// e.g. MakeActionInfo([]string{http.MethodGet, http.MethodPost})
func MakeActionInfo(httpMethods ...string) []hexapolicy.ActionInfo {
	actionInfos := make([]hexapolicy.ActionInfo, 0)
	for _, aMethod := range httpMethods {
		if aMethod == "" {
			// if testing with a "" action, pass it as is
			actionInfos = append(actionInfos, hexapolicy.ActionInfo{ActionUri: aMethod})
		} else {
			actionInfos = append(actionInfos, hexapolicy.ActionInfo{ActionUri: "http:" + aMethod})
		}
	}
	return actionInfos
}
