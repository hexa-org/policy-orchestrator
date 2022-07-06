package amazonwebservices

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"log"
	"net/http"
	"strings"
)

type CognitoClient interface {
	ListUserPools(ctx context.Context, params *cognitoidentityprovider.ListUserPoolsInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUserPoolsOutput, error)
	ListUsers(ctx context.Context, params *cognitoidentityprovider.ListUsersInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUsersOutput, error)
	AdminEnableUser(ctx context.Context, params *cognitoidentityprovider.AdminEnableUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminEnableUserOutput, error)
	AdminDisableUser(ctx context.Context, params *cognitoidentityprovider.AdminDisableUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminDisableUserOutput, error)
}

type AmazonProvider struct {
	CognitoClientOverride CognitoClient
}

func (a *AmazonProvider) Name() string {
	return "amazon"
}

func (a *AmazonProvider) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	if !strings.EqualFold(info.Name, a.Name()) {
		return []orchestrator.ApplicationInfo{}, nil
	}
	return a.ListUserPools(info)
}

func (a *AmazonProvider) ListUserPools(info orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	client, clientErr := a.getHttpClient(info)
	if clientErr != nil {
		return nil, clientErr
	}
	poolsInput := cognitoidentityprovider.ListUserPoolsInput{MaxResults: 20}
	pools, listErr := client.ListUserPools(context.Background(), &poolsInput)
	if listErr != nil {
		return nil, listErr
	}
	for _, p := range pools.UserPools {
		apps = append(apps, orchestrator.ApplicationInfo{
			ObjectID:    aws.ToString(p.Id),
			Name:        aws.ToString(p.Name),
			Description: "Cognito identity provider user pool",
		})
	}
	return apps, err
}

func (a *AmazonProvider) GetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	client, err := a.getHttpClient(integrationInfo)
	if err != nil {
		return nil, err
	}

	filter := "status=\"Enabled\""
	userInput := cognitoidentityprovider.ListUsersInput{UserPoolId: &applicationInfo.ObjectID, Filter: &filter}
	users, err := client.ListUsers(context.Background(), &userInput)
	if err != nil {
		return nil, err
	}
	members := a.membersFrom(users)

	var policies []policysupport.PolicyInfo
	policies = append(policies, policysupport.PolicyInfo{
		Meta:    policysupport.MetaInfo{Version: "0.5"},
		Actions: []policysupport.ActionInfo{{"aws:amazon.cognito/access"}}, // todo - not sure what this should be just yet.
		Subject: policysupport.SubjectInfo{Members: members},
		Object: policysupport.ObjectInfo{
			ResourceID: applicationInfo.ObjectID,
			Resources:  []string{applicationInfo.ObjectID},
		},
	})
	return policies, nil
}

func (a *AmazonProvider) SetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	validate := validator.New() // todo - move this up?
	errApp := validate.Struct(applicationInfo)
	if errApp != nil {
		return 0, errApp
	}
	errPolicies := validate.Var(policyInfos, "omitempty,dive")
	if errPolicies != nil {
		return 0, errPolicies
	}

	client, err := a.getHttpClient(integrationInfo)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	for _, policyInfo := range policyInfos {
		var newUsers []string
		for _, user := range policyInfo.Subject.Members {
			newUsers = append(newUsers, user)
		}

		filter := "status=\"Enabled\""
		userInput := cognitoidentityprovider.ListUsersInput{UserPoolId: &applicationInfo.ObjectID, Filter: &filter}
		users, listUsersErr := client.ListUsers(context.Background(), &userInput)
		if listUsersErr != nil {
			log.Println("Unable to find amazon cognito users.")
			return http.StatusInternalServerError, listUsersErr
		}
		existingUsers := a.membersFrom(users)

		enableErr := a.EnableUsers(client, applicationInfo.ObjectID, a.ShouldEnable(existingUsers, newUsers))
		if enableErr != nil {
			log.Println("Unable to enable amazon cognito users.")
			return http.StatusInternalServerError, enableErr
		}

		disable := a.DisableUsers(client, applicationInfo.ObjectID, a.ShouldDisable(existingUsers, newUsers))
		if disable != nil {
			log.Println("Unable to disable amazon cognito users.")
			return http.StatusInternalServerError, disable
		}
	}
	return http.StatusCreated, nil
}

func (a *AmazonProvider) membersFrom(users *cognitoidentityprovider.ListUsersOutput) []string {
	var members []string
	for _, u := range users.Users {
		for _, attr := range u.Attributes {
			if aws.ToString(attr.Name) == "email" {
				members = append(members, fmt.Sprintf("%s:%s", aws.ToString(u.Username), aws.ToString(attr.Value)))
			}
		}
	}
	return members
}

func (a *AmazonProvider) EnableUsers(client CognitoClient, userPoolId string, shouldEnable []string) error {
	for _, enable := range shouldEnable {
		enable := cognitoidentityprovider.AdminEnableUserInput{UserPoolId: &userPoolId, Username: &strings.Split(enable, ":")[0]}
		_, err := client.AdminEnableUser(context.Background(), &enable)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AmazonProvider) ShouldEnable(existingUsers []string, desiredUsers []string) []string {
	var shouldEnable []string
	for _, newUser := range desiredUsers {
		var contains = false
		for _, existingUser := range existingUsers {
			if newUser == existingUser {
				contains = true
			}
		}
		if !contains {
			shouldEnable = append(shouldEnable, newUser)
		}
	}
	return shouldEnable
}

func (a *AmazonProvider) DisableUsers(client CognitoClient, userPoolId string, shouldDisable []string) error {
	for _, disableUser := range shouldDisable {

		disable := cognitoidentityprovider.AdminDisableUserInput{UserPoolId: &userPoolId, Username: &strings.Split(disableUser, ":")[0]}
		_, err := client.AdminDisableUser(context.Background(), &disable)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AmazonProvider) ShouldDisable(existingUsers []string, desiredUsers []string) []string {
	var shouldDisable []string
	for _, existingUser := range existingUsers {
		var contains = false
		for _, newUser := range desiredUsers {
			if strings.Contains(newUser, existingUser) {
				contains = true
			}
		}
		if !contains {
			shouldDisable = append(shouldDisable, existingUser)
		}
	}
	return shouldDisable
}

type CredentialsInfo struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

func (a *AmazonProvider) Credentials(key []byte) CredentialsInfo {
	var foundCredentials CredentialsInfo
	_ = json.NewDecoder(bytes.NewReader(key)).Decode(&foundCredentials)
	return foundCredentials
}

func (a *AmazonProvider) getHttpClient(info orchestrator.IntegrationInfo) (CognitoClient, error) {
	if a.CognitoClientOverride != nil {
		return a.CognitoClientOverride, nil
	}

	foundCredentials := a.Credentials(info.Key)
	defaultConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{AccessKeyID: foundCredentials.AccessKeyID, SecretAccessKey: foundCredentials.SecretAccessKey},
		}),
		config.WithRegion(foundCredentials.Region),
	)
	if err != nil {
		return nil, err
	}
	return cognitoidentityprovider.NewFromConfig(defaultConfig), nil
}
