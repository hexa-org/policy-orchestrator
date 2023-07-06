package armmodel_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewApimServiceInfo(t *testing.T) {
	aFQId := fqId()
	serviceUrl := "https://someserviceurl"
	res := armmodel.NewApimServiceInfo(aFQId, resTypeApi, resName, resDisplayName, serviceUrl)
	assert.Equal(t, aFQId, res.FullyQualifiedId)
	assert.Equal(t, resGroup, res.ResourceGroup)
	assert.Equal(t, resName, res.Name)
	assert.Equal(t, resTypeApi, res.Type)
	assert.Equal(t, resDisplayName, res.DisplayName)
	assert.Equal(t, serviceUrl, res.ServiceUrl)
}
