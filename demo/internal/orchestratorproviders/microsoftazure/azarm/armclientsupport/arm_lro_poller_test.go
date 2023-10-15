package armclientsupport_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armclientsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestNewArmLroPoller_PanicsWithNil(t *testing.T) {
	assert.Panics(t, func() {
		_ = armclientsupport.NewArmLroPoller[someApiResponse](nil)
	})
}

func TestPollForResult_PollerError(t *testing.T) {
	aFunc := func() (*runtime.Poller[string], error) {
		return nil, errors.New("some-error")
	}

	poller := armclientsupport.NewArmLroPoller[string](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.ErrorContains(t, err, "some-error")
	assert.Empty(t, result)
}

func TestPollForResult_WithFirstResponseError(t *testing.T) {
	httpClient := testsupport.NewMockHTTPClient()
	url := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		expStatus:  http.StatusBadRequest,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)
	poller := armclientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "the operation failed or was cancelled")
	assert.Empty(t, result)
}

func TestPollForResult_WithNextResponseError(t *testing.T) {
	httpClient := testsupport.NewMockHTTPClient()
	url := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    url,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()

	httpClient.AddRequest(http.MethodGet, url, http.StatusBadRequest, nil)
	poller := armclientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "400")
	assert.Empty(t, result)
}

func TestPollForResult_WithPolling(t *testing.T) {
	httpClient := testsupport.NewMockHTTPClient()
	url := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    url,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)
	poller := armclientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, apiResult, result)
}

func TestPollForResult_WithPollingMultipleTimes(t *testing.T) {
	httpClient := testsupport.NewMockHTTPClient()
	url := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    url,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	httpClient.AddRequest(http.MethodGet, url, http.StatusAccepted, nil)
	httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)
	poller := armclientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, apiResult, result)
}

func TestPollForResult_WithPollingRetryAfter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long running test")
	}

	httpClient := testsupport.NewMockHTTPClient()
	url := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    url,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
		retryAfter: 1,
	}
	aFunc := funcBuilder.pollerFunc()

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)
	poller := armclientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, apiResult, result)
}

func TestAzurePoller(t *testing.T) {
	httpClient := testsupport.NewMockHTTPClient()

	url := "https://azure.stratatest.io/poller"
	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)

	firstResp := &http.Response{
		StatusCode: http.StatusAccepted,
		Header: http.Header{
			"Location":    []string{url},
			"Retry-After": []string{"1"},
		},
		Body: io.NopCloser(http.NoBody),
	}

	opts := &policy.ClientOptions{
		Retry:     policy.RetryOptions{MaxRetries: -1},
		Transport: httpClient,
	}
	pipeline := runtime.NewPipeline("testmodule", "v0.1.0", runtime.PipelineOptions{}, opts)
	lro, err := runtime.NewPoller[someApiResponse](firstResp, pipeline, nil)
	assert.NoError(t, err)

	var respFromCtx *http.Response
	ctxWithResp := runtime.WithCaptureResponse(context.Background(), &respFromCtx)

	out, err := lro.PollUntilDone(ctxWithResp, &runtime.PollUntilDoneOptions{Frequency: time.Millisecond})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respFromCtx.StatusCode)
	assert.Equal(t, apiResult, out)
}
