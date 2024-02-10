package amazonwebservices

import (
	"net/http"
	"strings"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/amazonwebservices/awscognito"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/amazonwebservices/awscommon"

	"github.com/go-playground/validator/v10"
)

type AmazonProvider struct {
	AwsClientOpts awscommon.AWSClientOptions
}

func (a *AmazonProvider) Name() string {
	return "amazon"
}

func (a *AmazonProvider) DiscoverApplications(info policyprovider.IntegrationInfo) ([]policyprovider.ApplicationInfo, error) {
	if !strings.EqualFold(info.Name, a.Name()) {
		return []policyprovider.ApplicationInfo{}, nil
	}

	client, err := awscognito.NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return nil, err
	}
	return client.ListUserPools()
}

func (a *AmazonProvider) GetPolicyInfo(info policyprovider.IntegrationInfo, applicationInfo policyprovider.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	client, err := awscognito.NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return nil, err
	}

	groups, err := client.GetGroups(applicationInfo.ObjectID)
	if err != nil {
		return nil, err
	}

	var policies []hexapolicy.PolicyInfo
	for groupName := range groups {
		members, err := client.GetMembersAssignedTo(applicationInfo, groupName)
		if err != nil {
			return nil, err
		}
		policies = append(policies, hexapolicy.PolicyInfo{
			Meta:    hexapolicy.MetaInfo{Version: "0.5"},
			Actions: []hexapolicy.ActionInfo{{groupName}},
			Subject: hexapolicy.SubjectInfo{Members: members},
			Object: hexapolicy.ObjectInfo{
				ResourceID: applicationInfo.Name,
			},
		})
	}

	return policies, nil
}

func (a *AmazonProvider) SetPolicyInfo(info policyprovider.IntegrationInfo, applicationInfo policyprovider.ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (int, error) {
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
