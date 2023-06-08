package amazonwebservices

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AmazonProvider struct {
	AwsClientOpts AWSClientOptions
}

func (a *AmazonProvider) Name() string {
	return "amazon"
}

func (a *AmazonProvider) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	if !strings.EqualFold(info.Name, a.Name()) {
		return []orchestrator.ApplicationInfo{}, nil
	}

	client, err := NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return nil, err
	}
	return client.listUserPools()
}

func (a *AmazonProvider) GetPolicyInfo(info orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	client, err := NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return nil, err
	}

	groups, err := client.getGroups(applicationInfo.ObjectID)
	if err != nil {
		return nil, err
	}

	var policies []policysupport.PolicyInfo
	for groupName := range groups {
		pol, err := client.convertGroupToPolicy(applicationInfo, groupName)
		if err != nil {
			return nil, err
		}
		policies = append(policies, pol)
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

	client, err := NewCognitoClient(info.Key, a.AwsClientOpts)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	allGroups, err := client.getGroups(applicationInfo.ObjectID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	for _, pol := range policyInfos {
		groupName := pol.Actions[0].ActionUri
		_, exists := allGroups[groupName]
		if !exists {
			continue
		}

		existingUserEmailMap, err := client.listUsersInGroup(groupName, applicationInfo.ObjectID)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		policyUserEmailMap := make(map[string]string)
		for _, mem := range pol.Subject.Members {
			memEmail := strings.Split(mem, ":")[1]
			userName, err := client.getPrincipalIdFromEmail(applicationInfo, memEmail)
			if err != nil {
				log.Println("Error getPrincipalIdFromEmail with email=", memEmail, " Error=", err)
				continue
			}
			policyUserEmailMap[userName] = memEmail
		}

		toRemove := findElementsNotExistsIn(existingUserEmailMap, policyUserEmailMap)
		err = client.removeUsersFromGroup(applicationInfo, groupName, toRemove)
		if err != nil {
			log.Println("Error removing users from group", groupName, "Error=", err)
			return http.StatusInternalServerError, err
		}

		toAdd := findElementsNotExistsIn(policyUserEmailMap, existingUserEmailMap)
		err = client.addUsersToGroup(applicationInfo, groupName, toAdd)
		if err != nil {
			log.Println("Error adding user to group", groupName, "Error=", err)
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusCreated, nil
}
