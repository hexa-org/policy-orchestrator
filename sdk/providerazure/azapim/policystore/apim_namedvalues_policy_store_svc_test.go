package policystore_test

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/policystore"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const azureKey = `
{
  "appId": "c8bc3028-510a-41d1-b1ee-0567a2174ec6",
  "secret": "2mu8Q~aCHhjoHjhhvKk02MJ-3OMNgnkdh6YdibEs",
  "tenant": "d9908771-1b32-4795-ab43-46d107b75a53",
  "subscription": "f2f21609-3ca6-40dc-9a2d-511d705c49f5"
}
`

func TestGetPolicies(t *testing.T) {
	h := &http.Client{}
	//svcClient, err := client.NewNamedValuesClient([]byte(azureKey), h)
	//assert.NoError(t, err)
	svc, err := policystore.NewNamedValuePolicyStoreSvc([]byte(azureKey), h, "canarybankv2", "canarybankapi")
	assert.NoError(t, err)
	rarList, err := svc.GetPolicies(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, rarList)
	for _, aRar := range rarList {
		fmt.Println(aRar.Resource(), aRar.Actions(), aRar.Members())
	}
}

func TestSetPolicies(t *testing.T) {
	h := &http.Client{}
	//svcClient, err := client.NewNamedValuesClient([]byte(azureKey), h)
	//assert.NoError(t, err)
	svc, err := policystore.NewNamedValuePolicyStoreSvc([]byte(azureKey), h, "canarybankv2", "canarybankapi")
	assert.NoError(t, err)
	aRar, err := rar.NewResourceActionRoles("/analytics", []string{http.MethodGet}, []string{"Read.Analytics"})
	assert.NoError(t, err)
	err = svc.SetPolicy(aRar)
	fmt.Println(err)
}
