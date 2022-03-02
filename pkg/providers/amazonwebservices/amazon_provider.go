package amazonwebservices

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"strings"
)

type CognitoClient interface {
	ListUserPools(ctx context.Context, params *cognitoidentityprovider.ListUserPoolsInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUserPoolsOutput, error)
	ListUsers(ctx context.Context, params *cognitoidentityprovider.ListUsersInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUsersOutput, error)
}

type AmazonProvider struct {
	Client CognitoClient
}

func (a *AmazonProvider) Name() string {
	return "amazon"
}

func (a *AmazonProvider) DiscoverApplications(info provider.IntegrationInfo) ([]provider.ApplicationInfo, error) {
	if strings.EqualFold(info.Name, a.Name()) {
		err := a.ensureClientIsAvailable(info)
		if err != nil {
			return nil, err
		}
		return a.ListUserPools()
	}
	return []provider.ApplicationInfo{}, nil
}

func (a *AmazonProvider) ListUserPools() (apps []provider.ApplicationInfo, err error) {
	poolsInput := cognitoidentityprovider.ListUserPoolsInput{MaxResults: 20}
	pools, err := a.Client.ListUserPools(context.Background(), &poolsInput)
	if err != nil {
		return nil, err
	}
	for _, p := range pools.UserPools {
		apps = append(apps, provider.ApplicationInfo{
			ObjectID:    aws.ToString(p.Id),
			Name:        aws.ToString(p.Name),
			Description: "Cognito identity provider user pool",
		})
	}
	return apps, err
}

func (a *AmazonProvider) GetPolicyInfo(info provider.IntegrationInfo, info2 provider.ApplicationInfo) ([]provider.PolicyInfo, error) {
	userInput := cognitoidentityprovider.ListUsersInput{UserPoolId: &info2.ObjectID}
	users, err := a.Client.ListUsers(context.Background(), &userInput)
	if err != nil {
		return nil, err
	}

	var authenticatedUsers []string
	for _, u := range users.Users {
		for _, attr := range u.Attributes {
			if aws.ToString(attr.Name) == "email" {
				authenticatedUsers = append(authenticatedUsers, aws.ToString(attr.Value))
			}
		}
	}

	var policies []provider.PolicyInfo
	policies = append(policies, provider.PolicyInfo{
		Version: "0.3",
		Action:  "Access", // todo - not sure what this should be just yet.
		Subject: provider.SubjectInfo{AuthenticatedUsers: authenticatedUsers},
		Object:  provider.ObjectInfo{Resources: []string{info2.ObjectID}},
	})
	return policies, nil
}

func (a *AmazonProvider) SetPolicyInfo(info provider.IntegrationInfo, info2 provider.ApplicationInfo, info3 provider.PolicyInfo) error {
	return nil
}

type CredentialsInfo struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

func (g *AmazonProvider) Credentials(key []byte) CredentialsInfo {
	var foundCredentials CredentialsInfo
	_ = json.NewDecoder(bytes.NewReader(key)).Decode(&foundCredentials)
	return foundCredentials
}

func (a *AmazonProvider) ensureClientIsAvailable(info provider.IntegrationInfo) error {
	foundCredentials := a.Credentials(info.Key)
	if a.Client == nil {
		defaultConfig, err := config.LoadDefaultConfig(context.Background(),
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{AccessKeyID: foundCredentials.AccessKeyID, SecretAccessKey: foundCredentials.SecretAccessKey},
			}),
			config.WithRegion(foundCredentials.Region),
		)
		if err != nil {
			return err
		}
		a.Client = cognitoidentityprovider.NewFromConfig(defaultConfig)
	}
	return nil
}
