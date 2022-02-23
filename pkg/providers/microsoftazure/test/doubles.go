package microsoftazure_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
)

type MockExchange struct {
	Path         string
	Err          error
	ResponseBody []byte
	RequestBody  []byte
}

type MockClient struct {
	mock.Mock
	Exchanges []MockExchange
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	for _, e := range m.Exchanges {
		if e.Path == req.URL.String() {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(e.ResponseBody))}, e.Err
		}
	}
	return nil, errors.New(fmt.Sprintf("Missing mock exchange for %s.", req.RequestURI))
}

func (m *MockClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	for _, e := range m.Exchanges {
		if e.Path == url {
			e.RequestBody, _ = io.ReadAll(body)
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(e.ResponseBody))}, e.Err
		}
	}
	return nil, errors.New(fmt.Sprintf("Missing mock exchange for %s.", url))
}
