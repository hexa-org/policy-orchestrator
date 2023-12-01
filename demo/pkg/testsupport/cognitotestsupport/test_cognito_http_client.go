package cognitotestsupport

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
	"log"
	"net/http"
)

type MockCognitoHTTPClient struct {
	mock.Mock
	responseBody map[string][][]byte
	requestBody  map[string][][]byte
	statusCodes  map[string][]int
	called       map[string][]int
}

func NewMockCognitoHTTPClient() *MockCognitoHTTPClient {
	return &MockCognitoHTTPClient{
		Mock:         mock.Mock{},
		responseBody: make(map[string][][]byte),
		requestBody:  make(map[string][][]byte),
		statusCodes:  make(map[string][]int),
		called:       make(map[string][]int),
	}
}

func (m *MockCognitoHTTPClient) Do(req *http.Request) (*http.Response, error) {
	cognitoServiceOp := req.Header.Get("X-Amz-Target")
	reqKey := m.reqKey(req.Method, req.URL.String(), cognitoServiceOp)
	reqNum := len(m.called[reqKey])

	for expReqKey, expBodyList := range m.responseBody {
		if reqKey == expReqKey && reqNum < len(expBodyList) {
			return m.sendRequest(req.Method, req.URL.String(), cognitoServiceOp, req.Body)
		}
	}

	return nil, fmt.Errorf("missing mock response for request - %s Request Num %d", reqKey, reqNum)
}

func (m *MockCognitoHTTPClient) Get(url string) (resp *http.Response, err error) {
	return m.sendRequest(http.MethodGet, url, "", nil)
}

func (m *MockCognitoHTTPClient) Post(url, _ string, body io.Reader) (resp *http.Response, err error) {
	return m.sendRequest(http.MethodPost, url, "", body)
}

func (m *MockCognitoHTTPClient) sendRequest(method, url, cognitoServiceOp string, body io.Reader) (resp *http.Response, err error) {
	reqKey := m.reqKey(method, url, cognitoServiceOp)
	if body != nil {
		reqBody, _ := io.ReadAll(body)
		m.requestBody[reqKey] = append(m.requestBody[reqKey], reqBody)
	}

	reqNum := len(m.called[reqKey])
	var responseBody []byte
	responseBody = m.responseBody[reqKey][reqNum]
	statusCode := m.statusCodes[reqKey][reqNum]
	m.called[reqKey] = append(m.called[reqKey], statusCode)
	return &http.Response{StatusCode: statusCode, Body: io.NopCloser(bytes.NewReader(responseBody))}, nil
}

func (m *MockCognitoHTTPClient) AddRequest(method, url, apiOp string, statusCode int, responseBody []byte) {
	serviceOp := "AWSCognitoIdentityProviderService." + apiOp
	m.addRequest(m.reqKey(method, url, serviceOp), statusCode, responseBody)
}

func (m *MockCognitoHTTPClient) addRequest(reqKey string, statusCode int, responseBody []byte) {
	m.statusCodes[reqKey] = append(m.statusCodes[reqKey], statusCode)

	body := responseBody
	if responseBody == nil {
		body = make([]byte, 0)
	}
	m.responseBody[reqKey] = append(m.responseBody[reqKey], body)
}

func (m *MockCognitoHTTPClient) GetRequestBody(method, url, serviceOp string) []byte {
	return m.GetRequestBodyByIndex(method, url, serviceOp, 0)
}

func (m *MockCognitoHTTPClient) GetRequestBodyByIndex(method, url, serviceOp string, reqIndex int) []byte {
	reqKey := m.reqKey(method, url, serviceOp)
	if reqIndex < len(m.requestBody[reqKey]) {
		return m.requestBody[reqKey][reqIndex]
	}
	return nil
}

func (m *MockCognitoHTTPClient) reqKey(method, url, cognitoServiceOp string) string {
	if cognitoServiceOp != "" {
		return fmt.Sprintf("%s %s %s", method, url, cognitoServiceOp)
	}
	return method + " " + url
}

func (m *MockCognitoHTTPClient) VerifyCalled() bool {
	failCount := 0
	for reqKey, expStatusCodes := range m.statusCodes {
		expCount := len(expStatusCodes)
		calledCount := len(m.called[reqKey])
		if expCount == calledCount {
			continue
		}

		log.Println("Expected request not called. Request=", reqKey, "Counts: expected=", expCount, "called=", calledCount)
		failCount++

	}
	return failCount == 0
}
