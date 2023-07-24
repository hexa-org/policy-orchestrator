package awsapigw

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscognito"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"net/http"
)

type AwsApiGatewayProviderService struct {
	cognitoClient awscognito.CognitoClient
}

func NewAwsApiGatewayProviderService(cognitoClient awscognito.CognitoClient) *AwsApiGatewayProviderService {
	return &AwsApiGatewayProviderService{cognitoClient: cognitoClient}
}

func (s *AwsApiGatewayProviderService) DiscoverApplications(_ orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	return s.cognitoClient.ListUserPools()
}

func (s *AwsApiGatewayProviderService) GetPolicyInfo(appInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	groups, err := s.cognitoClient.GetGroups(appInfo.ObjectID)
	if err != nil {
		return nil, err
	}

	var policies []policysupport.PolicyInfo
	for groupName := range groups {
		members, err := s.cognitoClient.GetMembersAssignedTo(appInfo, groupName)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policysupport.PolicyInfo{
			Meta:    policysupport.MetaInfo{Version: "0.5"},
			Actions: []policysupport.ActionInfo{{groupName}},
			Subject: policysupport.SubjectInfo{Members: members},
			Object: policysupport.ObjectInfo{
				ResourceID: appInfo.Name,
			},
		})
	}

	return policies, nil
}

func (s *AwsApiGatewayProviderService) SetPolicyInfo(appInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
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
