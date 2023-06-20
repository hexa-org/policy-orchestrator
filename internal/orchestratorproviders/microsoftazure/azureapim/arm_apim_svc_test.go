package azureapim_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewArmApimSvc(t *testing.T) {
	svc, err := azureapim.NewArmApimSvc("", nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestGetResourceRoles_NoServiceInfo(t *testing.T) {
	svc, _ := azureapim.NewArmApimSvc("", nil, nil)
	roles, err := svc.GetResourceRoles(armmodel.ApimServiceInfo{})
	assert.NoError(t, err)
	assert.NotNil(t, roles)
	assert.Equal(t, []armmodel.ResourceActionRoles{}, roles)
}
