package google_cloud_test

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type MockClient struct {
	mock.Mock
	Err          error
	ResponseBody map[string][]byte
	RequestBody  []byte
}

func NewMockClient() *MockClient {
	return &MockClient{
		Mock:         mock.Mock{},
		Err:          nil,
		ResponseBody: make(map[string][]byte),
		RequestBody:  nil,
	}
}

func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	var body []byte
	if strings.Contains(url, "compute") {
		body = m.ResponseBody["compute"]
	} else {
		body = m.ResponseBody["appengine"]
	}
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(body))}, m.Err
}

func (m *MockClient) Post(url, _ string, body io.Reader) (resp *http.Response, err error) {
	m.RequestBody, _ = io.ReadAll(body)
	var responseBody []byte
	if strings.Contains(url, "compute") {
		responseBody = m.ResponseBody["compute"]
	} else {
		responseBody = m.ResponseBody["appengine"]
	}
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(responseBody))}, m.Err
}
