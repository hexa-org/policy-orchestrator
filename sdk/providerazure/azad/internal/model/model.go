package model

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func ToAppInfoList(result models.ApplicationCollectionResponseable) []idp.AppInfo {
	retApps := make([]idp.AppInfo, 0)
	for _, aResult := range result.GetValue() {
		retApps = append(retApps, ToAppInfo(aResult))
	}
	return retApps
}

func ToAppInfo(aResult models.Applicationable) idp.AppInfo {
	id := *aResult.GetId()
	appId := *aResult.GetAppId()
	name := *aResult.GetDisplayName()
	var aUri string
	if len(aResult.GetIdentifierUris()) > 0 {
		aUri = aResult.GetIdentifierUris()[0]
	}
	return NewResourceServerAppInfo(id, appId, name, aUri)
}
