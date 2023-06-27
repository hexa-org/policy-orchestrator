package azapim_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewArmApimSvc(t *testing.T) {
	svc, err := azapim.NewArmApimSvc("", nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestGetResourceRoles_NoServiceInfo(t *testing.T) {
	svc, _ := azapim.NewArmApimSvc("", nil, nil)
	roles, err := svc.GetResourceRoles(armmodel.ApimServiceInfo{})
	assert.NoError(t, err)
	assert.NotNil(t, roles)
	assert.Equal(t, []providerscommon.ResourceActionRoles{}, roles)
}
