package armmodel

import (
	"errors"
	"fmt"
	log "golang.org/x/exp/slog"
	"strings"
)

const ArmResourceSep = "/"

// ArmResource
// ResourceGroup - /subscriptions/<sub_id>/resourceGroups/<group>
type ArmResource struct {
	FullyQualifiedId string
	ResourceGroup    string
	Type             string
	Name             string
	DisplayName      string
}

func NewArmResource(fullyQualifiedId, resType, resName, displayName string) (ArmResource, error) {
	log.Info("NewArmResource", "fullId", fullyQualifiedId, "resType", resType, "resName", resName)
	fqId := strings.TrimPrefix(fullyQualifiedId, ArmResourceSep)
	parts := strings.Split(fqId, ArmResourceSep)
	log.Info("NewArmResource", "parts", parts)
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

	log.Info("NewArmResource 2", "parts", parts)
	return ArmResource{
		FullyQualifiedId: fullyQualifiedId,
		ResourceGroup:    parts[3],
		Type:             resType,
		Name:             resName,
		DisplayName:      displayName,
	}, nil
}
