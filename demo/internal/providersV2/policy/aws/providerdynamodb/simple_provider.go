package providerdynamodb

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	log "golang.org/x/exp/slog"
)

const AwsPolicyStoreTableName = "ResourcePolicies"

// resourcePolicyItem - definition of item stored in dynamodb
// meta tags are required
// This item specifies
// "Resource" column of type string, as the primary key
// "Action" column of type string, as the sort key
// "Members" column of type string, non-key
type resourcePolicyItem struct {
	Resource string `json:"Resource" meta:"resource,pk"`
	Action   string `json:"Action" meta:"actions,sk"`
	Members  string `json:"Members" meta:"members"`
}

func (t resourcePolicyItem) MapTo() (rar.ResourceActionRoles, error) {
	log.Info("resourcePolicyItem.MapTo", "msg", "Mapping", "rar", fmt.Sprintf("%v", t))
	members := make([]string, 0)
	err := json.Unmarshal([]byte(t.Members), &members)
	if err != nil {
		log.Error("resourcePolicyItem.MapTo", "msg", "Failed to unmarshal members string",
			"members", t.Members,
			"Err", err)
		return rar.ResourceActionRoles{}, err
	}
	return rar.NewResourceActionRoles(t.Resource, []string{t.Action}, members)
}
