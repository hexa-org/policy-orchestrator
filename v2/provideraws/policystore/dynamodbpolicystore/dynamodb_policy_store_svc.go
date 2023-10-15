package dynamodbpolicystore

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/v2/core/idp"
	"github.com/hexa-org/policy-orchestrator/v2/core/policyprovider"
	"github.com/hexa-org/policy-orchestrator/v2/core/policystore"
	"github.com/hexa-org/policy-orchestrator/v2/provideraws/awscommon"
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
	svc, err := NewPolicyStoreSvc(new(resourcePolicyItem), []byte("key"))
	app := new(idp.AppInfo)
	_, err = svc.GetPolicies(*app)
	if err != nil {
		return
	}
}

// Example end

type PolicyStoreSvc[T ItemInfo] struct {
	client   DynamodbClient
	itemType T
}

type Opt[T ItemInfo] func(svc *PolicyStoreSvc[T])

func WithDynamodbClientOverride[T ItemInfo](client DynamodbClient) Opt[T] {
	return func(svc *PolicyStoreSvc[T]) {
		svc.client = client
	}
}

func NewPolicyStoreSvc[T ItemInfo](itemType T, key []byte, opts ...Opt[T]) (policystore.PolicyStoreSvc, error) {
	if len(opts) == 0 {
		client, err := NewDynamodbClient(key, nil)
		if err != nil {
			return nil, err
		}
		return &PolicyStoreSvc[T]{client: client, itemType: itemType}, nil
	}

	svc := &PolicyStoreSvc[T]{itemType: itemType}
	for _, o := range opts {
		o(svc)
	}
	return svc, nil
}

func (s *PolicyStoreSvc[T]) GetPolicies(_ idp.AppInfo) ([]policyprovider.ResourceActionRoles, error) {
	tableName := "AwsPolicyStoreTableName"
	input := &ddb.ScanInput{TableName: &tableName}
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

func toResourceActionRoleList[T ItemInfo](items []T) []policyprovider.ResourceActionRoles {
	rarList := make([]policyprovider.ResourceActionRoles, 0)
	for _, item := range items {
		rarList = append(rarList, item.ResourceActionRole())
	}
	return rarList
}

func (s *PolicyStoreSvc[T]) SetPolicy(_ policyprovider.ResourceActionRoles) error {
	return nil
}

// DynamodbClient -
// BEGIN - copied from dynamodb_client.go
type DynamodbClient interface {
	Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error)
	UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error)
}

type dynamodbClient struct {
	internal *ddb.Client
}

// NewDynamodbClient - builds DynamodbClient with provide credentials and optional httpClient
// pass an httpClient to use for tests
func NewDynamodbClient(key []byte, httpClient awscommon.AWSHttpClient) (DynamodbClient, error) {
	cfg, err := awscommon.GetAwsClientConfig(key, httpClient)
	if err != nil {
		return nil, err
	}

	return &dynamodbClient{internal: ddb.NewFromConfig(cfg)}, nil
}

func (c *dynamodbClient) Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error) {
	return c.internal.Scan(ctx, params, optFns...)
}

func (c *dynamodbClient) UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error) {
	return c.internal.UpdateItem(ctx, params, optFns...)
}

// END copied from dynamodb_client.go
