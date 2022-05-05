package openpolicyagent_test

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
)

type MockClient struct {
	mock.Mock
	Request  []byte
	Response []byte
	Err      error
	Bundle   bytes.Buffer
}

func (m *MockClient) Get(_ string) (resp *http.Response, err error) {
	recorder := httptest.ResponseRecorder{}
	recorder.Body = bytes.NewBuffer(m.Response)
	return recorder.Result(), m.Err
}

func (m *MockClient) Post(_, contentType string, body io.Reader) (resp *http.Response, err error) {
	index := strings.Index(contentType, "=")
	boundary := contentType[index+1:]
	reader := multipart.NewReader(body, boundary)
	form, err := reader.ReadForm(2 << 20)
	bundle := form.File["bundle"][0]
	file, err := bundle.Open()
	m.Request, _ = io.ReadAll(file)

	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(m.Response))}, m.Err
}
