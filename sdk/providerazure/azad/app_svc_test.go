package azad_test

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azad"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azad/internal/client"
	"github.com/stretchr/testify/assert"
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

func TestMe(t *testing.T) {
	svcClient, err := client.NewAzureGraphClient([]byte(azureKey), nil)
	assert.NoError(t, err)
	svc := azad.NewAppInfoSvc(svcClient)
	applications, err := svc.GetApplications()
	assert.NoError(t, err)
	assert.NotEmpty(t, applications)
	fmt.Println(applications)
}
