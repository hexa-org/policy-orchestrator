package model

import (
	"errors"
	"fmt"
	logger "golang.org/x/exp/slog"
	"strings"
)

const ArmResourceSep = "/"

// ArmResource
// ResourceGroup - /subscriptions/<sub_id>/resourceGroups/<group>
type ArmResource struct {
	fullyQualifiedId string
	resourceGroup    string
	resourceType     string
	name             string
	displayName      string
}

func (ar *ArmResource) Id() string {
	return ar.fullyQualifiedId
}

func (ar *ArmResource) ResourceGroup() string {
	return ar.resourceGroup
}

func (ar *ArmResource) Type() string {
	return ar.resourceType
}

func (ar *ArmResource) Name() string {
	return ar.name
}

func (ar *ArmResource) DisplayName() string {
	return ar.displayName
}

func NewArmResource(fullyQualifiedId, resType, resName, displayName string) (*ArmResource, error) {
	logger.Info("NewArmResource", "fullId", fullyQualifiedId, "resType", resType, "resName", resName)
	fqId := strings.TrimPrefix(fullyQualifiedId, ArmResourceSep)
	parts := strings.Split(fqId, ArmResourceSep)
	logger.Info("NewArmResource", "parts", parts)
	if len(parts) < 4 {
		return nil, errors.New("invalid resource id, require mim 4 parts separated by '/'. Found=" + fullyQualifiedId)
	}

	validateParts := []string{"subscriptions", "resourceGroups"}
	for i, key := range validateParts {
		p := i * 2
		if parts[p] != key || parts[p+1] == "" {
			return nil,
				fmt.Errorf("invalid resource id. missing %s. fullyQualifiedId=%s", key, fullyQualifiedId)
		}
	}

	logger.Info("NewArmResource 2", "parts", parts)
	return &ArmResource{
		fullyQualifiedId: fullyQualifiedId,
		resourceGroup:    parts[3],
		resourceType:     resType,
		name:             resName,
		displayName:      displayName,
	}, nil
}
