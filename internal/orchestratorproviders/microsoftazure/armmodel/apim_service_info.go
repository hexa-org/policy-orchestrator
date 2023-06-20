package armmodel

import "golang.org/x/exp/slog"

// ApimServiceInfo
//
//	/subscriptions/f2f21609-3ca6-40dc-9a2d-511d705c49f5/resourceGroups/canarybankv2/providers/Microsoft.ApiManagement/service/canarybankapi
type ApimServiceInfo struct {
	ArmResource
	//Id          string
	//Name        string
	DisplayName string
	ServiceUrl  string
	//Type        string
}

func NewApimServiceInfo(fullyQualifiedId, resType, name, displayName, serviceUrl string) ApimServiceInfo {
	slog.Info("NewApimServiceInfo", "fullId", fullyQualifiedId, "resType", resType, "name", name, "serviceUrl", serviceUrl)
	armRes, _ := NewArmResource(fullyQualifiedId, resType, name, displayName)

	return ApimServiceInfo{
		ArmResource: armRes,
		ServiceUrl:  serviceUrl,
	}
	//return &ApimServiceInfo{Id: id, Name: name, DisplayName: displayName, ServiceUrl: serviceUrl, Type: Type}
}

/*func NewApimServiceInfo(id string, name string, displayName string, serviceUrl string, resType string) *ApimServiceInfo {
	return &ApimServiceInfo{
		ArmResource: ArmResource{
			FullyQualifiedId: id,
			Name:             name,
			ResourceGroup:    "",
			Type:             resType,
		},
		DisplayName: "",
		ServiceUrl:  "",
	}
	return &ApimServiceInfo{Id: id, Name: name, DisplayName: displayName, ServiceUrl: serviceUrl, Type: Type}
}
*/
