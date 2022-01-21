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
	RequestBody []byte
}

func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	if m.Err != nil {
		return resp, m.Err
	}
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(m.ResponseBody))}, nil
}

func (m *MockClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	if m.Err != nil {
		return resp, m.Err
	}
	m.RequestBody, _ = io.ReadAll(body)
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(m.ResponseBody))}, nil
}
