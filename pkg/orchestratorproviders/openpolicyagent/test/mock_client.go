package openpolicyagent_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Request  []byte
	Response []byte
	Status   int
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
	if m.Status == -1 {
		panic("shoot")
	}
	return &http.Response{StatusCode: m.Status, Body: ioutil.NopCloser(bytes.NewReader(m.Response))}, m.Err
}
