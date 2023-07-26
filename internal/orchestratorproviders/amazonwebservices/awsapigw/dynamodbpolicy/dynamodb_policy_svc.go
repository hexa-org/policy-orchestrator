package dynamodbpolicy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	log "golang.org/x/exp/slog"
)

var AwsPolicyStoreTableName = "ResourcePolicies"

type PolicyStoreSvc interface {
	GetResourceRoles() ([]providerscommon.ResourceActionRoles, error)
	UpdateResourceRole(rar providerscommon.ResourceActionRoles) error
}

type policyStoreSvc struct {
	client DynamodbClient
}

func NewPolicyStoreSvc(client DynamodbClient) PolicyStoreSvc {
	return &policyStoreSvc{client: client}
}

func (p *policyStoreSvc) GetResourceRoles() ([]providerscommon.ResourceActionRoles, error) {
	input := &dynamodb.ScanInput{TableName: &AwsPolicyStoreTableName}
	output, err := p.client.Scan(context.TODO(), input)

	if err != nil {
		log.Error("PolicyStoreSvc.GetResourceRoles", "Failed to Scan table. Err=", err)
		return nil, err
	}

	if output == nil || len(output.Items) == 0 {
		log.Error("PolicyStoreSvc.GetResourceRoles", "Scan returned nil output or empty results. Output=", output)
		return nil, errors.New("scan returned nil output or empty results")
	}

	log.Info("PolicyStoreSvc.GetResourceRoles", "Scan result len=", len(output.Items))

	var items []resourcePolicyItem

	err = attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetResourceRoles", "Failed to unmarshal items. Err=", err)
		return nil, err
	}

	return toResourceActionRoleList(items)
}

func (p *policyStoreSvc) UpdateResourceRole(rar providerscommon.ResourceActionRoles) error {
	return p.UpdateResourceRole(rar)
}

type resourcePolicyItem struct {
	Resource string `json:"Resource"`
	Action   string `json:"Action"`
	Members  string `json:"Members"`
}

func toResourceActionRoleList(items []resourcePolicyItem) ([]providerscommon.ResourceActionRoles, error) {
	rarList := make([]providerscommon.ResourceActionRoles, 0)
	for _, item := range items {
		one, err := item.toResourceActionRole()
		if err != nil {
			return nil, err
		}
		rarList = append(rarList, one)
	}
	return rarList, nil
}

func (item resourcePolicyItem) toResourceActionRole() (providerscommon.ResourceActionRoles, error) {
	members := make([]string, 0)
	err := json.Unmarshal([]byte(item.Members), &members)
	if err != nil {
		log.Error("PolicyStoreSvc.toResourceActionRole", "Failed to unmarshal members string", item.Members, "Err", err)
		return providerscommon.ResourceActionRoles{}, err
	}
	return providerscommon.NewResourceActionRoles(item.Resource, item.Action, members), nil
}

/*
PolicyStoreSvcOpt
type PolicyStoreSvcOpt func(s *policyStoreSvc)

func WithDynamoDbClient(client DynamodbClient) PolicyStoreSvcOpt {
	return func(s *policyStoreSvc) {
		s.client = client
	}
}
*/
