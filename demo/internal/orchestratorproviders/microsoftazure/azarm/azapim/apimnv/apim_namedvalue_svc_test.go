package apimnv_test

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/providerscommon"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetResourceRoles_NoServiceInfo(t *testing.T) {
	svc := apimnv.NewApimNamedValueSvc("", nil, nil)
	roles, err := svc.GetResourceRoles(armmodel.ApimServiceInfo{})
	assert.NoError(t, err)
	assert.NotNil(t, roles)
	assert.Equal(t, []providerscommon.ResourceActionRoles{}, roles)
}
