package testsupport

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
)

type MockHTTPClient struct {
	mock.Mock
	Err          error
	ResponseBody map[string][]byte
	RequestBody  map[string][]byte
	Url          string
	StatusCode   int
	StatusCodes  map[string]int
	Called       map[string]int
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Mock:         mock.Mock{},
		Err:          nil,
		ResponseBody: make(map[string][]byte),
		RequestBody:  make(map[string][]byte),
		Url:          "",
		StatusCode:   200,
		StatusCodes:  make(map[string]int),
		Called:       make(map[string]int),
	}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {

	cognitoServiceOp := req.Header.Get("X-Amz-Target")
	methodAndUrl := m.reqKey(req.Method, req.URL.String(), cognitoServiceOp)

	for url := range m.ResponseBody {
		if url == req.URL.String() || url == methodAndUrl {
			return m.sendRequest(req.Method, req.URL.String(), cognitoServiceOp, req.Body)
		}
	}

	return nil, fmt.Errorf("missing mock response for request - " + methodAndUrl)
}

func (m *MockHTTPClient) Get(url string) (resp *http.Response, err error) {
	return m.sendRequest(http.MethodGet, url, "", nil)
}

func (m *MockHTTPClient) Post(url, _ string, body io.Reader) (resp *http.Response, err error) {
	return m.sendRequest(http.MethodPost, url, "", body)
}

func (m *MockHTTPClient) sendRequest(method, url, cognitoServiceOp string, body io.Reader) (resp *http.Response, err error) {
	m.Url = url
	statusCode := m.StatusCode
	reqKey := url

	methodAndUrl := m.reqKey(method, url, cognitoServiceOp)
	if code, exists := m.StatusCodes[methodAndUrl]; exists {
		reqKey = methodAndUrl
		statusCode = code
	}

	if body != nil {
		reqBody, _ := io.ReadAll(body)
		m.RequestBody[reqKey] = reqBody
	}

	var responseBody []byte
	responseBody = m.ResponseBody[reqKey]
	m.Called[reqKey] = statusCode
	return &http.Response{StatusCode: statusCode, Body: io.NopCloser(bytes.NewReader(responseBody))}, m.Err
}

func (m *MockHTTPClient) AddRequest(method, url string, statusCode int, responseBody []byte) {
	reqKey := m.reqKey(method, url, "")
	m.addRequest(reqKey, statusCode, responseBody)
}

func (m *MockHTTPClient) AddCognitoRequest(method, url, apiOp string, statusCode int, responseBody []byte) {
	serviceOp := "AWSCognitoIdentityProviderService." + apiOp
	m.addRequest(m.reqKey(method, url, serviceOp), statusCode, responseBody)
}

func (m *MockHTTPClient) addRequest(reqKey string, statusCode int, responseBody []byte) {
	m.StatusCodes[reqKey] = statusCode

	body := responseBody
	if responseBody == nil {
		body = make([]byte, 0)
	}
	m.ResponseBody[reqKey] = body
}

func (m *MockHTTPClient) GetRequestBody(url string) []byte {
	return m.GetRequestBodyByKey("", url)
}

func (m *MockHTTPClient) GetRequestBodyByKey(method, url string) []byte {
	reqKey := method + " " + url
	if _, exists := m.StatusCodes[reqKey]; exists {
		return m.RequestBody[reqKey]
	}
	return m.RequestBody[url]
}

func (m *MockHTTPClient) GetCognitoRequestBody(method, url, serviceOp string) []byte {
	reqKey := m.reqKey(method, url, serviceOp)
	if _, exists := m.StatusCodes[reqKey]; exists {
		return m.RequestBody[reqKey]
	}
	return m.RequestBody[url]
}

func (m *MockHTTPClient) CalledWithStatus(method, url string, expStatusCode int) bool {
	reqKey := method + " " + url
	actStatusCode, exists := m.Called[reqKey]
	if exists {
		return actStatusCode == expStatusCode
	}

	return m.Called[url] == expStatusCode
}

func (m *MockHTTPClient) VerifyCalled() bool {
	failCount := 0
	for reqKey := range m.StatusCodes {
		_, exists := m.Called[reqKey]
		if !exists {
			log.Println("Expected request not called. Request=", reqKey)
			failCount++
		}
	}
	return failCount == 0
}

func (m *MockHTTPClient) reqKey(method, reqUrl, cognitoServiceOp string) string {
	sortedUrl := sortedQueryUrl(reqUrl)
	if cognitoServiceOp != "" {
		return fmt.Sprintf("%s %s %s", method, sortedUrl, cognitoServiceOp)
	}
	return method + " " + sortedUrl
}

func sortedQueryUrl(origUrl string) string {
	sortedUrl, _ := url.Parse(origUrl)
	if len(sortedUrl.Query()) == 0 {
		return origUrl
	}

	sortedParamNames := make([]string, 0)
	for pName, _ := range sortedUrl.Query() {
		sortedParamNames = append(sortedParamNames, pName)
	}
	sort.Slice(sortedParamNames, func(i, j int) bool {
		return sortedParamNames[i] <= sortedParamNames[j]
	})

	sortedParams := url.Values{}
	for _, pName := range sortedParamNames {
		newVal := make([]string, 0)
		newVal = append(newVal, sortedUrl.Query()[pName]...)
		sort.Slice(newVal, func(i, j int) bool {
			return newVal[i] <= newVal[j]
		})
		sortedParams[pName] = newVal
	}

	sortedUrl.RawQuery = ""
	sortedUrl.ForceQuery = true
	return sortedUrl.String() + sortedParams.Encode()
}
