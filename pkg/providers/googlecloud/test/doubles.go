package google_cloud_test

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
)

type MockClient struct {
	mock.Mock
	Err          error
	ResponseBody []byte
	RequestBody  []byte
}

func (m *MockClient) Get(_ string) (resp *http.Response, err error) {
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(m.ResponseBody))}, m.Err
}

func (m *MockClient) Post(_, _ string, body io.Reader) (resp *http.Response, err error) {
	m.RequestBody, _ = io.ReadAll(body)
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(m.ResponseBody))}, m.Err
}
