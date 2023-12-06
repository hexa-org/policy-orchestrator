package model_test

import (
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

const gatewayUrl = "https://someserviceurl"

func TestNewArmApiAppInfo(t *testing.T) {
	aFQId := fqId()
	anApp := model.NewArmApiAppInfo(aFQId, resTypeApi, resName, resDisplayName, gatewayUrl)
	assert.Equal(t, resGroup, anApp.Id())
	assert.Equal(t, resName, anApp.Name())
	assert.Equal(t, "Azure APIM Service", anApp.Type())
	assert.Equal(t, resDisplayName, anApp.DisplayName())
	assert.Equal(t, gatewayUrl, anApp.GatewayUrl())
}
