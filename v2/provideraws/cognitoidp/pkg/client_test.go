package cognitoidp

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/v2/provideraws/app/cognitoidp/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListUserPools(t *testing.T) {
	client := newClient()
	pools, err := client.listUserPools()
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, pools)
}

func TestListResourceServers(t *testing.T) {
	httptest.NewRequest()
	client := newClient()
	pools, err := client.listResourceServers(testhelper.TestUserPoolId)
	assert.ErrorContains(t, err, "StatusCode: 400")
	fmt.Println(err)
	assert.Nil(t, pools)
}

func newClient() CognitoClient {
	httpClient := &http.Client{Timeout: time.Second}
	client, _ := NewCognitoClient(testhelper.AwsCredentialsForTest(), httpClient)
	return client
}
