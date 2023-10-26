package dynamodbpolicystore

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/client"
	log "golang.org/x/exp/slog"
)

type PolicyStoreSvc[R rar.ResourceActionRolesMapper] struct {
	client    client.DynamodbClient
	tableInfo TableInfo[R]
}

type Opt[R rar.ResourceActionRolesMapper] func(svc *PolicyStoreSvc[R])

func WithDynamodbClientOverride[R rar.ResourceActionRolesMapper](client client.DynamodbClient) Opt[R] {
	return func(svc *PolicyStoreSvc[R]) {
		svc.client = client
	}
}

func NewPolicyStoreSvc[R rar.ResourceActionRolesMapper](tableInfo TableInfo[R], key []byte, opts ...Opt[R]) (policystore.PolicyBackendSvc[R], error) {
	svc := &PolicyStoreSvc[R]{tableInfo: tableInfo}
	if len(opts) == 0 {
		c, err := client.NewDynamodbClient(key, nil)
		if err != nil {
			return nil, err
		}
		svc.client = c
	}

	for _, o := range opts {
		o(svc)
	}
	return svc, nil
}

func (s *PolicyStoreSvc[R]) GetPolicies(_ idp.AppInfo) ([]rar.ResourceActionRoles, error) {
	input := &ddb.ScanInput{TableName: &s.tableInfo.TableName}
	output, err := s.client.Scan(context.TODO(), input)

	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to Scan table. Err=", err)
		return nil, err
	}

	var items []R
	err = attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		log.Error("PolicyStoreSvc.GetPolicies", "Failed to unmarshal items. Err=", err)
		return nil, err
	}

	return rar.ToResourceActionRoleList(items)
}

func (s *PolicyStoreSvc[R]) SetPolicy(rar rar.ResourceActionRoles) error {
	tableDefinition := s.tableInfo.TableDefinition
	inputBuilder := client.NewInputBuilder(s.tableInfo.TableName, map[string]string{
		"Resource": tableDefinition.ResourceAttrName,
		"Action":   tableDefinition.ActionAttrName,
		"Members":  tableDefinition.MembersAttrName,
	})

	input, err := inputBuilder.UpdateItemInput(rar)
	if err != nil {
		return err
	}

	// TODO - process output
	_, err = s.client.UpdateItem(context.TODO(), input)
	return err
}
