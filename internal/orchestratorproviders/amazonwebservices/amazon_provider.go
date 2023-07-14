package amazonwebservices

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscognito"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AmazonProvider struct {
	AwsClientOpts awscommon.AWSClientOptions
}

func (a *AmazonProvider) Name() string {
	return "amazon"
}

func (a *AmazonProvider) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	if !strings.EqualFold(info.Name, a.Name()) {
		return []orchestrator.ApplicationInfo{}, nil
	}

	client, err := awscognito.NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return nil, err
	}
	return client.ListUserPools()
}

func (a *AmazonProvider) GetPolicyInfo(info orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	client, err := awscognito.NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return nil, err
	}

	groups, err := client.GetGroups(applicationInfo.ObjectID)
	if err != nil {
		return nil, err
	}

	var policies []policysupport.PolicyInfo
	for groupName := range groups {
		members, err := client.GetMembersAssignedTo(applicationInfo, groupName)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policysupport.PolicyInfo{
			Meta:    policysupport.MetaInfo{Version: "0.5"},
			Actions: []policysupport.ActionInfo{{groupName}},
			Subject: policysupport.SubjectInfo{Members: members},
			Object: policysupport.ObjectInfo{
				ResourceID: applicationInfo.Name,
			},
		})
	}

	return policies, nil
}

func (a *AmazonProvider) SetPolicyInfo(info orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	validate := validator.New() // todo - move this up?
	err := validate.Struct(applicationInfo)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = validate.Var(policyInfos, "omitempty,dive")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	client, err := awscognito.NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	allGroups, err := client.GetGroups(applicationInfo.ObjectID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	for _, pol := range policyInfos {
		groupName := pol.Actions[0].ActionUri
		_, exists := allGroups[groupName]
		if !exists {
			continue
		}

		err = client.SetGroupsAssignedTo(groupName, pol.Subject.Members, applicationInfo)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusCreated, nil
}
