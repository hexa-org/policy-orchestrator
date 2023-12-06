package model

import (
	logger "golang.org/x/exp/slog"
)

// ArmApiAppInfo - consider moving this to core
// same struct used by cognito
//"id": "/subscriptions/f2f21609-3ca6-40dc-9a2d-511d705c49f5/resourceGroups/canarybankv2/providers/Microsoft.ApiManagement/service/canarybankapi",
//"name": "canarybankapi",
//"type": "Microsoft.ApiManagement/service",
//"properties": { "gatewayUrl": "https://canarybankapi.azure-api.net", ... }

type ArmApiAppInfo struct {
	//*ArmResource
	resourceGroup string
	serviceName   string
	displayName   string
	gatewayUrl    string
}

func NewArmApiAppInfo(fullyQualifiedId, resType, name, displayName, gatewayUrl string) *ArmApiAppInfo {
	logger.Info("NewArmApiAppInfo", "fullId", fullyQualifiedId, "resType", resType, "name", name, "gatewayUrl", gatewayUrl)
	armRes, _ := NewArmResource(fullyQualifiedId, resType, name, displayName)

	return &ArmApiAppInfo{
		resourceGroup: armRes.resourceGroup,
		serviceName:   name,
		displayName:   displayName,
		gatewayUrl:    gatewayUrl,
	}
}

// Id - maps to ObjectID
func (a *ArmApiAppInfo) Id() string {
	return a.resourceGroup
}

// Name - maps to Name
func (a *ArmApiAppInfo) Name() string {
	return a.serviceName
}
func (a *ArmApiAppInfo) DisplayName() string {
	return a.displayName
}
func (a *ArmApiAppInfo) GatewayUrl() string {
	return a.gatewayUrl
}
func (a *ArmApiAppInfo) Type() string {
	// "type": "Microsoft.ApiManagement/service",
	return "Azure APIM Service"
}
