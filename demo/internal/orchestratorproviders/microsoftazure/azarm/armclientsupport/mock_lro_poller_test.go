package armclientsupport_test

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armclientsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"io"
	"net/http"
	"strconv"
	"testing"
)

type someApiResponse struct {
	Name string
}

type pollerFuncBuilder struct {
	httpClient *testsupport.MockHTTPClient
	nextUrl    string
	expStatus  int
	expBody    io.Reader
	retryAfter int
}

func (b pollerFuncBuilder) pollerFunc() armclientsupport.GetPollerFunc[someApiResponse] {
	opts := &policy.ClientOptions{
		Retry:     policy.RetryOptions{MaxRetries: -1},
		Transport: b.httpClient,
	}
	pipeline := runtime.NewPipeline("testmodule", "v0.1.0", runtime.PipelineOptions{}, opts)

	expStatus := b.expStatus
	headers := http.Header{}
	if b.nextUrl != "" {
		headers["Location"] = []string{b.nextUrl}
	}

	if !testing.Short() && b.retryAfter > 0 {
		headers["Retry-After"] = []string{strconv.Itoa(b.retryAfter)}
	}

	body := io.NopCloser(b.expBody)
	aFunc := func() (*runtime.Poller[someApiResponse], error) {
		firstResp := &http.Response{StatusCode: expStatus, Header: headers, Body: body}
		return runtime.NewPoller[someApiResponse](firstResp, pipeline, nil)
	}

	return aFunc
}

/*
func TestPollForResult_WithPolling(t *testing.T) {
	apiResp := someApiResponse{Name: "some name"}
	aFunc := func() (*runtime.Poller[someApiResponse], error) {
		respBytes, _ := json.Marshal(apiResp)
		resp := initialResponse(http.MethodGet, "", respBytes)
		return runtime.NewPoller(resp, runtime.Pipeline{}, &runtime.NewPollerOptions[someApiResponse]{
			Handler: &mockPollerHandler{
				numTimesToPoll: 2,
			},
		})
	}
	poller := armclientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(1)
	assert.NoError(t, err)
	assert.Equal(t, apiResp, result)
}

func initialResponse(method, u string, respBody []byte) *http.Response {
	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		panic(err)
	}
	return &http.Response{
		Body:          io.NopCloser(bytes.NewReader(respBody)),
		ContentLength: -1,
		Header:        http.Header{},
		Request:       req,
	}
}

type mockPollerHandler struct {
	curr           int
	numTimesToPoll int
	httpResp       *http.Response
}

func (mh *mockPollerHandler) Done() bool {
	return mh.curr == mh.numTimesToPoll
}

func (mh *mockPollerHandler) Poll(_ context.Context) (*http.Response, error) {
	var httpResp *http.Response
	mh.curr = mh.curr + 1
	if mh.curr < mh.numTimesToPoll {
		httpResp = &http.Response{StatusCode: http.StatusAccepted}
	} else {
		out := someApiResponse{Name: "some name"}
		respBytes, _ := json.Marshal(out)
		respBody := io.NopCloser(bytes.NewReader(respBytes))
		httpResp = &http.Response{StatusCode: http.StatusOK, Body: respBody}
	}

	mh.httpResp = httpResp
	return httpResp, nil
}

func (mh *mockPollerHandler) Result(_ context.Context, out *someApiResponse) error {
	body := mh.httpResp.Body
	resp, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, out)
	return err
}
*/
