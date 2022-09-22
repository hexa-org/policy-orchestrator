package google_cloud_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Err          error
	ResponseBody map[string][]byte
	RequestBody  []byte
	Url          string
}

func NewMockClient() *MockClient {
	return &MockClient{
		Mock:         mock.Mock{},
		Err:          nil,
		ResponseBody: make(map[string][]byte),
		RequestBody:  nil,
		Url:          "",
	}
}

func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	m.Url = url
	var body []byte
	if strings.Contains(url, "compute") {
		body = m.ResponseBody["compute"]
	} else {
		body = m.ResponseBody["appengine"]
	}
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(body))}, m.Err
}

func (m *MockClient) Post(url, _ string, body io.Reader) (resp *http.Response, err error) {
	m.Url = url
	m.RequestBody, _ = io.ReadAll(body)
	var responseBody []byte
	if strings.Contains(url, "compute") {
		responseBody = m.ResponseBody["compute"]
	} else {
		responseBody = m.ResponseBody["appengine"]
	}
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(responseBody))}, m.Err
}
