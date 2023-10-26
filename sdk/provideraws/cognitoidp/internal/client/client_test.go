package cognitoidp

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/app/cognitoidp/internal/testhelper"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon"
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
	client := newClient(m)
	act, err := client.ListUserPools()
	assert.NoError(t, err)
	assert.NotNil(t, act)
	assert.Equal(t, 1, len(act.UserPools))
	assert.Equal(t, testhelper.TestUserPoolName, *act.UserPools[0].Name)
	assert.Equal(t, testhelper.TestUserPoolId, *act.UserPools[0].Id)
}

// TestListUserPools_Error - asserts an error is returned when calling the real cognito
// listUserPools endpoint. Since this uses random aws credentials, an error is expected.
func TestListUserPools_Error(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}
	client := newClient(httpClient)
	pools, err := client.ListUserPools()
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, pools)
}

// TestListResourceServers_Error - asserts an error is returned calling the real cognito
// listResourceServers endpoint. Since this uses random aws credentials, an error is expected.
func TestListResourceServers_Error(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}
	client := newClient(httpClient)
	pools, err := client.ListResourceServers(testhelper.TestUserPoolId)
	assert.ErrorContains(t, err, "StatusCode: 400")
	fmt.Println(err)
	assert.Nil(t, pools)
}

func newClient(httpClient awscommon.AWSHttpClient) CognitoClient {
	client, _ := NewCognitoClient(testhelper.AwsCredentialsForTest(), httpClient)
	return client
}
