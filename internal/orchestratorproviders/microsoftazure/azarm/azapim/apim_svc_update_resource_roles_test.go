package azapim_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	apim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestUpdateResourceRole_ValidationErrors(t *testing.T) {
	svc, _ := azapim.NewArmApimSvc("", nil, nil)
	serviceInfo := armmodel.ApimServiceInfo{
		ArmResource: armmodel.ArmResource{ResourceGroup: "", Name: "service"},
	}
	err := svc.UpdateResourceRole(serviceInfo, providerscommon.ResourceActionRoles{})
	assert.Error(t, err)

	serviceInfo = armmodel.ApimServiceInfo{
		ArmResource: armmodel.ArmResource{ResourceGroup: "resGroup", Name: ""},
	}
	err = svc.UpdateResourceRole(serviceInfo, providerscommon.ResourceActionRoles{})
	assert.Error(t, err)
}

func TestUpdateResourceRole_Error(t *testing.T) {
	nvClient := armtestsupport.NewMockNamedValuesClient()
	svc := makeArmApiSvc(nvClient)
	serviceInfo := makeServiceInfo()

	existingNV := providerscommon.NewResourceActionRoles("/humanresources", http.MethodGet, []string{""})
	updateReqParams := makeNVClientUpdateParam(existingNV)
	var poller *runtime.Poller[apim.NamedValueClientUpdateResponse]
	nvClient.On("BeginUpdate", context.Background(), serviceInfo.ResourceGroup, serviceInfo.Name, existingNV.Name(), "*", updateReqParams, nil).
		Return(poller, errors.New("poller error"))

	err := svc.UpdateResourceRole(serviceInfo, existingNV)
	assert.ErrorContains(t, err, "poller error")
}

func TestUpdateResourceRole_Success(t *testing.T) {
	nvClient := armtestsupport.NewMockNamedValuesClient()
	svc := makeArmApiSvc(nvClient)

	serviceInfo := makeServiceInfo()

	existingNV := providerscommon.NewResourceActionRoles("/humanresources", http.MethodGet, []string{""})

	updatedNV := providerscommon.NewResourceActionRoles("/humanresources", http.MethodGet, []string{"GetHRRole"})
	expResp := makeNVClientResp(updatedNV)
	poller := makePoller(expResp)

	updateReqParams := makeNVClientUpdateParam(existingNV)
	nvClient.On("BeginUpdate", context.Background(), serviceInfo.ResourceGroup, serviceInfo.Name, existingNV.Name(), "*", updateReqParams, nil).
		Return(poller, nil)

	err := svc.UpdateResourceRole(serviceInfo, existingNV)
	assert.NoError(t, err)
}

func makeServiceInfo() armmodel.ApimServiceInfo {
	resGroup := armtestsupport.ApimResourceGroupName
	service := armtestsupport.ApimServiceName
	return armmodel.ApimServiceInfo{
		ArmResource: armmodel.ArmResource{
			ResourceGroup: resGroup,
			Name:          service,
		},
	}
}

func makeNVClientResp(nv providerscommon.ResourceActionRoles) apim.NamedValueClientUpdateResponse {
	nvName := nv.Name()
	nvVal := nv.Value()
	return apim.NamedValueClientUpdateResponse{
		NamedValueContract: apim.NamedValueContract{
			Properties: &apim.NamedValueContractProperties{
				Value: &nvVal,
			},
			Name: &nvName,
		},
	}
}

func makeNVClientUpdateParam(updatedNv providerscommon.ResourceActionRoles) apim.NamedValueUpdateParameters {
	nvVal := updatedNv.Value()
	return apim.NamedValueUpdateParameters{
		Properties: &apim.NamedValueUpdateParameterProperties{
			Value: &nvVal,
		},
	}
}

func makePoller(expResp apim.NamedValueClientUpdateResponse) *runtime.Poller[apim.NamedValueClientUpdateResponse] {
	respBytes, _ := json.Marshal(expResp)
	firstResp := initialResponse(http.MethodPatch, "", respBytes)
	firstResp.StatusCode = http.StatusOK
	poller, _ := runtime.NewPoller[apim.NamedValueClientUpdateResponse](firstResp, runtime.Pipeline{}, nil)
	return poller
}

func makeArmApiSvc(nvClient apimnv.NamedValuesClient) azapim.ArmApimSvc {
	opt := azapim.WithNamedValuesClient(nvClient)
	svc, _ := azapim.NewArmApimSvc("", nil, nil, opt)
	return svc
}

func initialResponse(method, reqUrl string, respBody []byte) *http.Response {
	req, _ := http.NewRequest(method, reqUrl, nil)

	return &http.Response{
		Body:          io.NopCloser(bytes.NewReader(respBody)),
		ContentLength: -1,
		Header:        http.Header{},
		Request:       req,
	}
}
