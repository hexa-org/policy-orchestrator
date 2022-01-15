package google_cloud_test

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
)

type MockClient struct {
	mock.Mock
	Err  error
	Json string
}

func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	_, file, _, _ := runtime.Caller(0)
	jsonFile := filepath.Join(file, "./../backends.json")
	readFile, _ := ioutil.ReadFile(jsonFile)
	if m.Err != nil {
		return resp, m.Err
	}
	if m.Json != "" {
		readFile = []byte(m.Json)
	}
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(readFile))}, nil
}
