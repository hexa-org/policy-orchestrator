package apps_test

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/apps"
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

func TestMe(t *testing.T) {
	h := &http.Client{}
	//svcClient, err := client.NewApimServiceClient([]byte(azureKey), h)
	//assert.NoError(t, err)
	svc, err := apps.NewAppInfoSvc([]byte(azureKey), h)
	applications, err := svc.GetApplications()
	assert.NoError(t, err)
	assert.NotEmpty(t, applications)
	app := applications[0]
	fmt.Println(app.Name(), app.Id(), app.DisplayName(), app.Type())
}
