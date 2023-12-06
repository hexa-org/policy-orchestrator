package clientsupport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/clientsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

type MockHttpClient struct {
	mock.Mock
	requests map[string]int
	loop     int
}

func NewMockHTTPClient() *MockHttpClient {
	return &MockHttpClient{requests: make(map[string]int)}
}

func reqKey(req *http.Request) string {
	return req.Method + "::" + req.URL.String()
}
func (m *MockHttpClient) addRequest(req *http.Request, resp *http.Response) {
	rKey := reqKey(req)
	count := m.requests[rKey]
	m.requests[rKey] = count + 1
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if m.loop > 10 {
		return nil, fmt.Errorf("MockHttpClient breaking out of possible infinite loop. Loop count=%d", m.loop)
	}

	m.addRequest(req, nil)
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
func (m *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	fmt.Println("arm_lro_poller_test.MockHttpClient", "GET", url)
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}
func (m *MockHttpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}
func (m *MockHttpClient) httpReqMatcher(expReq *http.Request, num int) interface{} {
	m.loop = m.loop + 1

	return mock.MatchedBy(func(req *http.Request) bool {
		fmt.Println("httpReqMatcher EXP", expReq.Method, expReq.URL.String())
		fmt.Println("httpReqMatcher ACTUAL", req.Method, req.URL.String())

		if req.Method != expReq.Method || req.URL.String() != expReq.URL.String() {
			fmt.Println("httpReqMatcher method OR URL did not match")
			return false
		}
		for name, val := range expReq.Header {
			if req.Header.Get(name) != val[0] {
				fmt.Println("httpReqMatcher header did not match", "name", name, "expVal", val[0])
				return false
			}
		}

		// poller makes multiple calls to same URL, keep track of current request
		// so we need to make sure we are matching against the correct request
		// otherwise this will go into an infinite loop
		rKey := reqKey(req)
		if m.requests[rKey] != num {
			fmt.Println("httpReqMatcher reqNum DOES NOT MATCH", "num", num, "m.reqNum", num)
			return false
		}

		fmt.Println("httpReqMatcher MATCHES TRUE")
		return true
	})
}

func TestNewArmLroPoller_PanicsWithNil(t *testing.T) {
	assert.Panics(t, func() {
		_ = clientsupport.NewArmLroPoller[someApiResponse](nil)
	})
}

func TestPollForResult_PollerError(t *testing.T) {
	aFunc := func() (*runtime.Poller[string], error) {
		return nil, errors.New("some-error")
	}

	poller := clientsupport.NewArmLroPoller[string](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.ErrorContains(t, err, "some-error")
	assert.Empty(t, result)
}

func TestPollForResult_WithFirstResponseError(t *testing.T) {
	httpClient := NewMockHTTPClient()
	reqUrl := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		expStatus:  http.StatusBadRequest,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	parsedUrl, _ := url.Parse(reqUrl)
	httpReq := &http.Request{Method: http.MethodGet, URL: parsedUrl}
	httpResp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(respBytes)), Request: httpReq}

	httpClient.On("Get", reqUrl).
		Return(httpResp, nil)

	//httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)
	poller := clientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "the operation failed or was cancelled")
	assert.Empty(t, result)
}

func TestPollForResult_WithNextResponseError(t *testing.T) {
	httpClient := NewMockHTTPClient()
	reqUrl := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    reqUrl,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()
	parsedUrl, _ := url.Parse(reqUrl)

	httpReq := &http.Request{Method: http.MethodGet, URL: parsedUrl}
	httpResp := &http.Response{StatusCode: http.StatusBadRequest,
		Body:    io.NopCloser(bytes.NewReader([]byte{})),
		Request: httpReq}

	httpClient.On("Do", httpClient.httpReqMatcher(httpReq, 0)).
		Return(httpResp, nil)
	poller := clientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "400")
	assert.Empty(t, result)
}

func TestPollForResult_WithPolling(t *testing.T) {
	httpClient := NewMockHTTPClient()
	reqUrl := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    reqUrl,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	//httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)
	parsedUrl, _ := url.Parse(reqUrl)
	httpReq := &http.Request{Method: http.MethodGet, URL: parsedUrl}
	httpResp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(respBytes)), Request: httpReq}

	httpClient.On("Do", httpClient.httpReqMatcher(httpReq, 0)).
		Return(httpResp, nil)
	poller := clientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, apiResult, result)
}

func TestPollForResult_WithPollingMultipleTimes(t *testing.T) {
	httpClient := NewMockHTTPClient()
	reqUrl := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    reqUrl,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
	}
	aFunc := funcBuilder.pollerFunc()

	parsedUrl, _ := url.Parse(reqUrl)
	httpReq := &http.Request{Method: http.MethodGet, URL: parsedUrl}
	httpResp := &http.Response{StatusCode: http.StatusAccepted, Body: http.NoBody}

	httpClient.On("Do", httpClient.httpReqMatcher(httpReq, 0)).
		Return(httpResp, nil)

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)
	//httpClient.AddRequest(http.MethodGet, url, http.StatusOK, respBytes)
	parsedUrl, _ = url.Parse(reqUrl)
	httpReq = &http.Request{Method: http.MethodGet, URL: parsedUrl}
	httpResp = &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(respBytes))}

	httpClient.On("Do", httpClient.httpReqMatcher(httpReq, 1)).
		Return(httpResp, nil)

	poller := clientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, apiResult, result)
}

func TestPollForResult_WithPollingRetryAfter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long running test")
	}

	httpClient := NewMockHTTPClient()
	reqUrl := "https://azure.stratatest.io/poller"
	funcBuilder := pollerFuncBuilder{
		httpClient: httpClient,
		nextUrl:    reqUrl,
		expStatus:  http.StatusAccepted,
		expBody:    http.NoBody,
		retryAfter: 1,
	}
	aFunc := funcBuilder.pollerFunc()

	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)

	parsedUrl, _ := url.Parse(reqUrl)
	httpReq := &http.Request{Method: http.MethodGet, URL: parsedUrl}
	httpResp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(respBytes)), Request: httpReq}

	httpClient.On("Do", httpClient.httpReqMatcher(httpReq, 0)).
		Return(httpResp, nil)
	poller := clientsupport.NewArmLroPoller[someApiResponse](aFunc)
	result, err := poller.PollForResult(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, apiResult, result)
}

func TestAzurePoller(t *testing.T) {
	httpClient := NewMockHTTPClient()

	reqUrl := "https://azure.stratatest.io/poller"
	apiResult := someApiResponse{Name: "some name"}
	respBytes, _ := json.Marshal(apiResult)

	parsedUrl, _ := url.Parse(reqUrl)
	httpReq := &http.Request{Method: http.MethodGet, URL: parsedUrl}
	httpResp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(respBytes)), Request: httpReq}

	httpClient.On("Do", httpClient.httpReqMatcher(httpReq, 0)).
		Return(httpResp, nil)

	firstResp := &http.Response{
		StatusCode: http.StatusAccepted,
		Header: http.Header{
			"Location":    []string{reqUrl},
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

type someApiResponse struct {
	Name string
}

type pollerFuncBuilder struct {
	httpClient *MockHttpClient
	nextUrl    string
	expStatus  int
	expBody    io.Reader
	retryAfter int
}

func (b pollerFuncBuilder) pollerFunc() clientsupport.GetPollerFunc[someApiResponse] {
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
