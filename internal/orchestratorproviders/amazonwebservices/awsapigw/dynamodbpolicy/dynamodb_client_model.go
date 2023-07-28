package dynamodbpolicy

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	log "golang.org/x/exp/slog"
	"strings"
)

func UpdateItemInput(rar providerscommon.ResourceActionRoles) (*dynamodb.UpdateItemInput, error) {
	return updateItemInput(rar.Resource, rar.Action, rar.Roles)
}

func updateItemInput(resource string, action string, members []string) (*dynamodb.UpdateItemInput, error) {
	if strings.TrimSpace(resource) == "" || strings.TrimSpace(action) == "" {
		return nil, fmt.Errorf("empty resource='%s' or action='%s'", resource, action)
	}

	aResource, _ := attributevalue.Marshal(strings.TrimSpace(resource))
	anAction, _ := attributevalue.Marshal(strings.TrimSpace(action))
	keyAttrVal := map[string]types.AttributeValue{"Resource": aResource, "Action": anAction}

	membersVal, err := membersAttributeValue(members)
	if err != nil {
		log.Error("updateItemInput error building AttributeValue from", "members", members, "Err", err)
		return nil, err
	}
	updateExpr := "SET #members = :members"
	input := &dynamodb.UpdateItemInput{TableName: &AwsPolicyStoreTableName,
		Key: keyAttrVal,
		ExpressionAttributeNames: map[string]string{
			"#members": "Members",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":members": membersVal,
		},
		UpdateExpression: &updateExpr,
		ReturnValues:     types.ReturnValueAllNew,
	}
	return input, nil
}

func membersAttributeValue(members []string) (types.AttributeValue, error) {
	sanitizedMembers := providerscommon.SanitizeMembers(members)
	membersStr, err := json.Marshal(sanitizedMembers)
	if err != nil {
		return nil, err
	}
	return attributevalue.Marshal(string(membersStr))
}
