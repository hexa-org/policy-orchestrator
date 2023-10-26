package dynamodbpolicystore

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policyprovider"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	log "golang.org/x/exp/slog"
)

type ItemInfo interface {
	ResourceActionRole() policyprovider.ResourceActionRoles
}

type DbInfo interface {
	TableName() string
}

// resourcePolicyItem
// Example begin
type resourcePolicyItem struct {
	Resource string `json:"Resource"`
	Action   string `json:"Action"`
	Members  string `json:"Members"`
}

func (it resourcePolicyItem) ResourceActionRole() policyprovider.ResourceActionRoles {
	members := make([]string, 0)
	_ = json.Unmarshal([]byte(it.Members), &members)
	return policyprovider.ResourceActionRoles{Resource: it.Resource, Action: it.Action, Roles: members}
}

func RunMe() {
	tableName := "AwsPolicyStoreTableName"
	svc, err := NewPolicyStoreSvc(new(resourcePolicyItem), []byte("key"), tableName)
	app := new(idp.AppInfo)
	_, err = svc.GetPolicies(*app)
	if err != nil {
		return
	}
}

// Example end

type PolicyStoreSvc[T ItemInfo] struct {
	client    DynamodbClient
	itemType  T
	tableName string
}

type Opt[T ItemInfo] func(svc *PolicyStoreSvc[T])

func WithDynamodbClientOverride[T ItemInfo](client DynamodbClient) Opt[T] {
	return func(svc *PolicyStoreSvc[T]) {
		svc.client = client
	}
}

func NewPolicyStoreSvc[T ItemInfo](itemType T, key []byte, tableName string, opts ...Opt[T]) (policystore.PolicyBackendSvc, error) {
	svc := &PolicyStoreSvc[T]{itemType: itemType, tableName: tableName}
	if len(opts) == 0 {
		client, err := NewDynamodbClient(key, nil)
		if err != nil {
			return nil, err
		}
		svc.client = client
	}

	for _, o := range opts {
		o(svc)
	}
	return svc, nil
}

func (s *PolicyStoreSvc[T]) GetPolicies(_ idp.AppInfo) ([]policyprovider.ResourceActionRoles, error) {
	input := &ddb.ScanInput{TableName: &s.tableName}
	output, err := s.client.Scan(context.TODO(), input)

	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to Scan table. Err=", err)
		return nil, err
	}

	var items []T
	err = attributevalue.UnmarshalListOfMaps(output.Items, items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to unmarshal items. Err=", err)
		return nil, err
	}

	return nil, nil
}

func (s *PolicyStoreSvc[T]) SetPolicy(_ policyprovider.ResourceActionRoles) error {
	return nil
}

func toResourceActionRoleList[T ItemInfo](items []T) []policyprovider.ResourceActionRoles {
	rarList := make([]policyprovider.ResourceActionRoles, 0)
	for _, item := range items {
		rarList = append(rarList, item.ResourceActionRole())
	}
	return rarList
}
