package testsupport

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockHTTPClient struct {
	mock.Mock
	Err          error
	ResponseBody map[string][]byte
	RequestBody  []byte
	Url          string
	StatusCode   int
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Mock:         mock.Mock{},
		Err:          nil,
		ResponseBody: make(map[string][]byte),
		RequestBody:  nil,
		Url:          "",
		StatusCode:   200,
	}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	for url, body := range m.ResponseBody {
		if url == req.URL.String() {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(body))}, m.Err
		}
	}
	return nil, fmt.Errorf("Missing mock response for %s.", req.URL.String())
}

func (m *MockHTTPClient) Get(url string) (resp *http.Response, err error) {
	m.Url = url
	var body []byte
	body = m.ResponseBody[url]
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body))}, m.Err
}

func (m *MockHTTPClient) Post(url, _ string, body io.Reader) (resp *http.Response, err error) {
	m.Url = url
	m.RequestBody, _ = io.ReadAll(body)
	var responseBody []byte
	responseBody = m.ResponseBody[url]
	return &http.Response{StatusCode: m.StatusCode, Body: io.NopCloser(bytes.NewReader(responseBody))}, m.Err
}
