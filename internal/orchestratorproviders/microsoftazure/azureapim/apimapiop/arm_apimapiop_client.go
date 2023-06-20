package apimapiop

import "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"

type ArmApimApiOpClient interface {
}

type armApimApiOpClient struct {
	internal *armapimanagement.APIOperationPolicyClient
}
