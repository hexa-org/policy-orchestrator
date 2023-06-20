package armmodel

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slog"
	"strings"
)

const ArmResourceSep = "/"

// ArmResource
//
//	ResourceGroup - /subscriptions/f2f21609-3ca6-40dc-9a2d-511d705c49f5/resourceGroups/canarybankv2
type ArmResource struct {
	FullyQualifiedId string
	ResourceGroup    string
	Type             string
	Name             string
	DisplayName      string
}

func NewArmResource(fullyQualifiedId, resType, resName, displayName string) (ArmResource, error) {
	slog.Info("NewArmResource", "fullId", fullyQualifiedId, "resType", resType, "resName", resName)
	fqId := strings.TrimPrefix(fullyQualifiedId, ArmResourceSep)
	parts := strings.Split(fqId, ArmResourceSep)
	slog.Info("NewArmResource", "parts", parts)
	if len(parts) < 4 {
		return ArmResource{}, errors.New("invalid resource id, require mim 4 parts separated by '/'. Found=" + fullyQualifiedId)
	}

	validateParts := []string{"subscriptions", "resourceGroups"}
	for i, key := range validateParts {
		p := i * 2
		if parts[p] != key || parts[p+1] == "" {
			return ArmResource{},
				fmt.Errorf("invalid resource id. missing %s. fullyQualifiedId=%s", key, fullyQualifiedId)
		}
	}

	slog.Info("NewArmResource 2", "parts", parts)
	return ArmResource{
		FullyQualifiedId: fullyQualifiedId,
		ResourceGroup:    parts[3],
		Type:             resType,
		Name:             resName,
		DisplayName:      displayName,
	}, nil
}

/*
func ParseFullyQualifiedId(fullyQualifiedId string) (map[string]string, error) {
	fqId := strings.TrimPrefix(fullyQualifiedId, ArmResourceSep)
	parts := strings.Split(fqId, ArmResourceSep)
	if len(parts) < 4 {
		return nil, errors.New("invalid resource id, require mim 4 parts separated by '/'. Found=" + fullyQualifiedId)
	}

	kvMap := make(map[string]string)
	i := 0
	for i < len(parts) {
		k := parts[i]
		v := parts[i+1]
		kvMap[k] = v
		i++
	}

	for _, key := range []string{"subscriptions", "resourceGroups"} {
		if val, found := kvMap[key]; !found || val == "" {
			return nil, fmt.Errorf("invalid resource id. missing %s. fullyQualifiedId=%s", key, fullyQualifiedId)
		}
	}

	return kvMap, nil
}

func NewArmResourceOld(fullyQualifiedId string) (ArmResource, error) {
	idParts := strings.Split(fullyQualifiedId, "/")
	armRes := ArmResource{}
	if len(idParts) != 9 {
		return armRes, errors.New("invalid resource id, require 8 parts separated by '/'. Found=" + fullyQualifiedId)
	}

	resourceGroup := idParts[4]
	serviceName := idParts[8]
	log.Println("NewArmResourceOld fullyQualifiedId=", fullyQualifiedId)
	log.Println("NewArmResourceOld resourceGroup=", resourceGroup, "serviceName=", serviceName)
	if serviceName == "" || resourceGroup == "" {
		return armRes, errors.New("invalid resource id. resourceGroup or serviceName is blank. fullyQualifiedId=" + fullyQualifiedId)
	}
	return ArmResource{
		FullyQualifiedId: fullyQualifiedId,
		Name:             serviceName,
		ResourceGroup:    resourceGroup,
	}, nil
}

*/
