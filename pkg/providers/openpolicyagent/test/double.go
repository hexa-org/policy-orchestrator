package openpolicyagent_test

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type MockClient struct {
	mock.Mock
	Request     []byte
	Response    []byte
	Err         error
}

func (m *MockClient) Get(_ string) (resp *http.Response, err error) {
	recorder := httptest.ResponseRecorder{}
	recorder.Body = bytes.NewBuffer(m.Response)
	return recorder.Result(), m.Err
}

func (m *MockClient) Post(_, _ string, body io.Reader) (resp *http.Response, err error) {
	m.Request, _ = io.ReadAll(body)
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(m.Response))}, m.Err
}
