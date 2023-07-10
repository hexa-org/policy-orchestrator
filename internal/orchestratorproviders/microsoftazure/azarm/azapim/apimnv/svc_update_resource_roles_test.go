package apimnv_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestUpdateResourceRole_ValidationErrors(t *testing.T) {
	svc := apimnv.NewApimNamedValueSvc("", nil, nil)
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
	nvClient := apim_testsupport.NewMockNamedValuesClient()
	svc := makeApimNamedValueSvc(nvClient)
	serviceInfo := makeServiceInfo()

	existingNV := providerscommon.NewResourceActionRoles("/humanresources", http.MethodGet, []string{""})
	updateReqParams := makeNVClientUpdateParam(existingNV)
	var poller *runtime.Poller[azarmapim.NamedValueClientUpdateResponse]
	nvClient.On("BeginUpdate", context.Background(), serviceInfo.ResourceGroup, serviceInfo.Name, existingNV.Name(), "*", updateReqParams, nil).
		Return(poller, errors.New("poller error"))

	err := svc.UpdateResourceRole(serviceInfo, existingNV)
	assert.ErrorContains(t, err, "poller error")
}

func TestUpdateResourceRole_Success(t *testing.T) {
	nvClient := apim_testsupport.NewMockNamedValuesClient()
	svc := makeApimNamedValueSvc(nvClient)

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

func makeNVClientResp(nv providerscommon.ResourceActionRoles) azarmapim.NamedValueClientUpdateResponse {
	nvName := nv.Name()
	nvVal := nv.Value()
	return azarmapim.NamedValueClientUpdateResponse{
		NamedValueContract: azarmapim.NamedValueContract{
			Properties: &azarmapim.NamedValueContractProperties{
				Value: &nvVal,
			},
			Name: &nvName,
		},
	}
}

func makeNVClientUpdateParam(updatedNv providerscommon.ResourceActionRoles) azarmapim.NamedValueUpdateParameters {
	nvVal := updatedNv.Value()
	return azarmapim.NamedValueUpdateParameters{
		Properties: &azarmapim.NamedValueUpdateParameterProperties{
			Value: &nvVal,
		},
	}
}

func makePoller(expResp azarmapim.NamedValueClientUpdateResponse) *runtime.Poller[azarmapim.NamedValueClientUpdateResponse] {
	respBytes, _ := json.Marshal(expResp)
	firstResp := initialResponse(http.MethodPatch, "", respBytes)
	firstResp.StatusCode = http.StatusOK
	poller, _ := runtime.NewPoller[azarmapim.NamedValueClientUpdateResponse](firstResp, runtime.Pipeline{}, nil)
	return poller
}

func makeApimNamedValueSvc(nvClient apimnv.NamedValuesClient) apimnv.ApimNamedValueSvc {
	opt := apimnv.WithNamedValuesClient(nvClient)
	svc := apimnv.NewApimNamedValueSvc("", nil, nil, opt)
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
