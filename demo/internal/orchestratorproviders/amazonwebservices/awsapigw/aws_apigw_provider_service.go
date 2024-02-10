package awsapigw

import (
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/amazonwebservices/awsapigw/dynamodbpolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/amazonwebservices/awscognito"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/providerscommon"
	log "golang.org/x/exp/slog"
	"net/http"
)

type AwsApiGatewayProviderService struct {
	cognitoClient awscognito.CognitoClient
	policySvc     dynamodbpolicy.PolicyStoreSvc
}

func NewAwsApiGatewayProviderService(cognitoClient awscognito.CognitoClient, policySvc dynamodbpolicy.PolicyStoreSvc) *AwsApiGatewayProviderService {
	return &AwsApiGatewayProviderService{cognitoClient: cognitoClient, policySvc: policySvc}
}

func (s *AwsApiGatewayProviderService) DiscoverApplications(_ orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	return s.cognitoClient.ListUserPools()
}

func (s *AwsApiGatewayProviderService) GetPolicyInfo(appInfo orchestrator.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	rarList, err := s.policySvc.GetResourceRoles()
	if err != nil {
		log.Error("AwsApiGatewayProviderService.GetPolicyInfo", "error calling GetResourceRoles App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return []hexapolicy.PolicyInfo{}, err
	}
	return providerscommon.BuildPolicies(rarList), nil
}

func (s *AwsApiGatewayProviderService) SetPolicyInfo(appInfo orchestrator.ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (int, error) {
	rarList, err := s.policySvc.GetResourceRoles()
	if err != nil {
		log.Error("AwsApiGatewayProviderService.SetPolicyInfo", "error calling GetResourceRoles App.Name", appInfo.Name, "identifierUrl[0]", appInfo.Service, "err=", err)
		return http.StatusBadGateway, err
	}

	rarUpdateList := providerscommon.CalcResourceActionRolesForUpdate(rarList, policyInfos)
	for _, rar := range rarUpdateList {
		err = s.policySvc.UpdateResourceRole(rar)
		if err != nil {
			return http.StatusBadGateway, err
		}
	}
	return http.StatusCreated, nil
}

func (s *AwsApiGatewayProviderService) setPolicyInfoOld(appInfo orchestrator.ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (int, error) {
	allGroups, err := s.cognitoClient.GetGroups(appInfo.ObjectID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	for _, pol := range policyInfos {
		groupName := pol.Actions[0].ActionUri
		_, exists := allGroups[groupName]
		if !exists {
			continue
		}

		err = s.cognitoClient.SetGroupsAssignedTo(groupName, pol.Subject.Members, appInfo)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusCreated, nil
}
