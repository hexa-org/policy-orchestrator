package client

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

// TestListUserPools - asserts the cognito listUserPools endpoint is called
// and that the return ListUserPoolsOutput struct is as expected.
func TestListUserPools(t *testing.T) {
	m := testhelper.NewMockCognitoHttpClient()
	m.ExpectListUserPools(nil)
	c := newClient(m)
	act, err := c.ListUserPools()
	assert.NoError(t, err)
	assert.NotNil(t, act)
	assert.Equal(t, 1, len(act.UserPools))
	assert.Equal(t, testhelper.TestUserPoolName, *act.UserPools[0].Name)
	assert.Equal(t, testhelper.TestUserPoolId, *act.UserPools[0].Id)
}

func TestListResourceServers(t *testing.T) {
	m := testhelper.NewMockCognitoHttpClient()
	c := newClient(m)
	m.ExpectListResourceServers(nil)
	act, err := c.ListResourceServers(testhelper.TestUserPoolId)
	assert.NoError(t, err)
	assert.NotNil(t, act)
	assert.Equal(t, 1, len(act.ResourceServers))
	assert.Equal(t, testhelper.TestUserPoolId, *act.ResourceServers[0].UserPoolId)
	assert.Equal(t, testhelper.TestResourceServerName, *act.ResourceServers[0].Name)
	assert.Equal(t, testhelper.TestResourceServerIdentifier, *act.ResourceServers[0].Identifier)
}

// TestListUserPools_Error - asserts an error is returned when calling the real cognito
// listUserPools endpoint. Since this uses random aws credentials, an error is expected.
func TestListUserPools_Error(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}
	c := newClient(httpClient)
	pools, err := c.ListUserPools()
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, pools)
}

// TestListResourceServers_Error - asserts an error is returned calling the real cognito
// listResourceServers endpoint. Since this uses random aws credentials, an error is expected.
func TestListResourceServers_Error(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}
	c := newClient(httpClient)
	pools, err := c.ListResourceServers(testhelper.TestUserPoolId)
	assert.ErrorContains(t, err, "StatusCode: 400")
	fmt.Println(err)
	assert.Nil(t, pools)
}

func newClient(httpClient awscommon.AWSHttpClient) CognitoClient {
	c, _ := NewCognitoClient(testhelper.AwsCredentialsForTest(), httpClient)
	return c
}
