package armmodel_test

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/stretchr/testify/assert"
	"testing"
)

const resGroup = "canarybankv2"
const resTypeApi = "Microsoft.ApiManagement/service"
const resName = "canarybankapi"
const resDisplayName = "CanaryBankAPI"

func fqId() string {
	return fmt.Sprintf("/subscriptions/<subid>/resourceGroups/%s/providers/Microsoft.ApiManagement/service/%s", resGroup, resName)
}

func TestNewArmResource_LessThanFourParts(t *testing.T) {
	aFQId := "/subscriptions/<subid>/resourceGroups"
	res, err := armmodel.NewArmResource(aFQId, "", "", "")
	assert.ErrorContains(t, err, "4 parts")
	assert.Empty(t, res)
}

func TestNewArmResource_WithoutSubscription(t *testing.T) {
	aFQId := "/subscriptions/resourceGroups/" + resGroup
	res, err := armmodel.NewArmResource(aFQId, "", "", "")
	assert.ErrorContains(t, err, "4 parts")
	assert.Empty(t, res)
}

func TestNewArmResource_WithoutValidSubscription(t *testing.T) {
	aFQId := "/invalid/<subid>/resourceGroups/" + resGroup
	res, err := armmodel.NewArmResource(aFQId, "", "", "")
	assert.ErrorContains(t, err, "missing subscriptions")
	assert.Empty(t, res)
}

func TestNewArmResource_WithEmptySubscription(t *testing.T) {
	aFQId := "/invalid//resourceGroups/" + resGroup
	res, err := armmodel.NewArmResource(aFQId, "", "", "")
	assert.ErrorContains(t, err, "missing subscriptions")
	assert.Empty(t, res)
}

func TestNewArmResource_WithoutResourceGroup(t *testing.T) {
	aFQId := "/subscriptions/<subid>/resourceGroups/"
	res, err := armmodel.NewArmResource(aFQId, "", "", "")
	assert.ErrorContains(t, err, "missing resourceGroups")
	assert.Empty(t, res)
}

func TestNewArmResource_ResourceGroup(t *testing.T) {
	aFQId := "/subscriptions/<subid>/resourceGroups/" + resGroup
	res, err := armmodel.NewArmResource(aFQId, "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, aFQId, res.FullyQualifiedId)
	assert.Equal(t, resGroup, res.ResourceGroup)
}

func TestNewArmResource_Success(t *testing.T) {
	aFQId := fqId()
	res, err := armmodel.NewArmResource(aFQId, resTypeApi, resName, resDisplayName)
	assert.NoError(t, err)
	assert.Equal(t, aFQId, res.FullyQualifiedId)
	assert.Equal(t, resGroup, res.ResourceGroup)
	assert.Equal(t, resName, res.Name)
	assert.Equal(t, resTypeApi, res.Type)
	assert.Equal(t, resDisplayName, res.DisplayName)
}
