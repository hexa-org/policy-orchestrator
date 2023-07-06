package armmodel

import log "golang.org/x/exp/slog"

// ApimServiceInfo
// /subscriptions/<sub_id>/resourceGroups/<group>/providers/Microsoft.ApiManagement/service/<service>
type ApimServiceInfo struct {
	ArmResource
	ServiceUrl string
}

func NewApimServiceInfo(fullyQualifiedId, resType, name, displayName, serviceUrl string) ApimServiceInfo {
	log.Info("NewApimServiceInfo", "fullId", fullyQualifiedId, "resType", resType, "name", name, "serviceUrl", serviceUrl)
	armRes, _ := NewArmResource(fullyQualifiedId, resType, name, displayName)

	return ApimServiceInfo{
		ArmResource: armRes,
		ServiceUrl:  serviceUrl,
	}
}
