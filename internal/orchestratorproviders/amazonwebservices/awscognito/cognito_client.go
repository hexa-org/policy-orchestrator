package awscognito

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"log"
	"strings"
)

type CognitoClient interface {
	ListUserPools() (apps []orchestrator.ApplicationInfo, err error)
	GetGroups(userPoolId string) (map[string]string, error)
	GetMembersAssignedTo(appInfo orchestrator.ApplicationInfo, groupName string) ([]string, error)
	SetGroupsAssignedTo(groupName string, members []string, applicationInfo orchestrator.ApplicationInfo) error
}

type cognitoClient struct {
	client *cognitoidentityprovider.Client
}

func NewCognitoClient(key []byte, opt awscommon.AWSClientOptions) (CognitoClient, error) {
	client, err := newCognitoClient(key, opt)
	if err != nil {
		return nil, err
	}
	return &cognitoClient{client: client}, nil
}

func newCognitoClient(key []byte, opts awscommon.AWSClientOptions) (*cognitoidentityprovider.Client, error) {
	cfg, err := awscommon.GetAwsClientConfig(key, opts)
	if err != nil {
		return nil, err
	}

	return cognitoidentityprovider.NewFromConfig(cfg), nil
}

func (c *cognitoClient) ListUserPools() (apps []orchestrator.ApplicationInfo, err error) {
	pools, listErr := c.listUserPools()
	if listErr != nil {
		return nil, listErr
	}
	for _, p := range pools.UserPools {
		rsOutput, err := c.listResourceServers(*p.Id)
		if err != nil {
			return nil, err
		}

		for _, rs := range rsOutput.ResourceServers {
			apps = append(apps, orchestrator.ApplicationInfo{
				ObjectID:    *rs.UserPoolId,
				Name:        *rs.Name,
				Description: "Cognito",
				Service:     *rs.Identifier,
			})
		}
	}
	return apps, err
}

func (c *cognitoClient) listUserPools() (*cognitoidentityprovider.ListUserPoolsOutput, error) {
	poolsInput := cognitoidentityprovider.ListUserPoolsInput{MaxResults: 20}
	pools, err := c.client.ListUserPools(context.Background(), &poolsInput)
	return pools, err
}

func (c *cognitoClient) listResourceServers(userPoolId string) (*cognitoidentityprovider.ListResourceServersOutput, error) {
	rsInput := cognitoidentityprovider.ListResourceServersInput{UserPoolId: &userPoolId, MaxResults: 10}
	rsOutput, err := c.client.ListResourceServers(context.Background(), &rsInput)
	return rsOutput, err
}

func (c *cognitoClient) GetGroups(userPoolId string) (map[string]string, error) {
	groupsInput := cognitoidentityprovider.ListGroupsInput{
		UserPoolId: aws.String(userPoolId),
	}
	output, err := c.client.ListGroups(context.Background(), &groupsInput)
	if err != nil {
		return nil, err
	}

	groups := make(map[string]string)
	for _, g := range output.Groups {
		groups[aws.ToString(g.GroupName)] = aws.ToString(g.Description)
	}

	return groups, nil
}

func (c *cognitoClient) GetMembersAssignedTo(appInfo orchestrator.ApplicationInfo, groupName string) ([]string, error) {
	tmpUserEmailMap, err := c.listUsersInGroup(groupName, appInfo.ObjectID)

	if err != nil {
		return nil, err
	}

	members := make([]string, 0)
	for _, email := range tmpUserEmailMap {
		members = append(members, fmt.Sprintf("user:%s", email))
	}
	return members, nil
}

func (c *cognitoClient) SetGroupsAssignedTo(groupName string, members []string, applicationInfo orchestrator.ApplicationInfo) error {
	existingUserEmailMap, err := c.listUsersInGroup(groupName, applicationInfo.ObjectID)
	if err != nil {
		return err
	}

	policyUserEmailMap := make(map[string]string)
	for _, mem := range members {
		memEmail := strings.Split(mem, ":")[1]
		userName, err := c.getPrincipalIdFromEmail(applicationInfo, memEmail)
		if err != nil {
			log.Println("Error getPrincipalIdFromEmail with email=", memEmail, " Error=", err)
			continue
		}
		policyUserEmailMap[userName] = memEmail
	}

	toRemove := findElementsNotExistsIn(existingUserEmailMap, policyUserEmailMap)
	err = c.removeUsersFromGroup(applicationInfo, groupName, toRemove)
	if err != nil {
		log.Println("Error removing users from group", groupName, "Error=", err)
		return err
	}

	toAdd := findElementsNotExistsIn(policyUserEmailMap, existingUserEmailMap)
	err = c.addUsersToGroup(applicationInfo, groupName, toAdd)
	if err != nil {
		log.Println("Error adding user to group", groupName, "Error=", err)
		return err
	}

	return nil
}

func (c *cognitoClient) listUsersInGroup(groupName, userPoolId string) (map[string]string, error) {
	input := cognitoidentityprovider.ListUsersInGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(userPoolId),
	}
	output, err := c.client.ListUsersInGroup(context.Background(), &input)
	if err != nil {
		return nil, err
	}

	userEmailList := workflowsupport.ProcessAsync[string, types.UserType](output.Users, func(user types.UserType) (string, error) {
		userInput := cognitoidentityprovider.AdminGetUserInput{
			UserPoolId: aws.String(userPoolId),
			Username:   user.Username,
		}

		userInfo, err := c.client.AdminGetUser(context.Background(), &userInput)
		if err != nil {
			log.Println("amazon_provider listUsersInGroup error calling AdminGetUser. error=", err)
			return "", err
		}

		for _, attr := range userInfo.UserAttributes {
			if aws.ToString(attr.Name) == "email" {
				return aws.ToString(user.Username) + "-:::-" + aws.ToString(attr.Value), nil
			}
		}

		return "", errors.New("email attribute not found for username " + aws.ToString(user.Username))
	})

	userEmailMap := make(map[string]string)
	for _, userEmail := range userEmailList {
		username, email, found := strings.Cut(userEmail, "-:::-")
		if !found {
			continue
		}
		userEmailMap[username] = email
	}

	return userEmailMap, nil
}

func (c *cognitoClient) getPrincipalIdFromEmail(appInfo orchestrator.ApplicationInfo, email string) (string, error) {
	filter := fmt.Sprintf("email=\"%s\"", email)
	listUserInput := cognitoidentityprovider.ListUsersInput{UserPoolId: &appInfo.ObjectID, Filter: &filter}
	users, err := c.client.ListUsers(context.Background(), &listUserInput)
	if err != nil {
		return "", err
	}
	if len(users.Users) == 0 {
		return "", errors.New("user not found for email=" + email)
	}

	return *users.Users[0].Username, nil
}

func (c *cognitoClient) addUsersToGroup(appInfo orchestrator.ApplicationInfo, groupName string, toAdd []string) error {
	for _, principalId := range toAdd {
		input := cognitoidentityprovider.AdminAddUserToGroupInput{
			GroupName:  &groupName,
			UserPoolId: &appInfo.ObjectID,
			Username:   &principalId,
		}

		_, err := c.client.AdminAddUserToGroup(context.Background(), &input)
		if err != nil {
			log.Println("Error adding user to group. User=", principalId, "group=", groupName, "Error=", err)
			return err
		}
	}
	return nil
}

func (c *cognitoClient) removeUsersFromGroup(appInfo orchestrator.ApplicationInfo, groupName string, toAdd []string) error {
	for _, principalId := range toAdd {
		input := cognitoidentityprovider.AdminRemoveUserFromGroupInput{
			GroupName:  &groupName,
			UserPoolId: &appInfo.ObjectID,
			Username:   &principalId,
		}

		_, err := c.client.AdminRemoveUserFromGroup(context.Background(), &input)
		if err != nil {
			log.Println("Error removing user from group. User=", principalId, "group=", groupName, "Error=", err)
			return err
		}
	}
	return nil
}

func findElementsNotExistsIn(elements, lookIn map[string]string) []string {
	var difference []string
	for existing := range elements {
		fromPolicy := lookIn[existing]
		if fromPolicy == "" {
			difference = append(difference, existing)
		}
	}
	return difference
}
