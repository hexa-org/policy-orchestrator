package amazonsupport_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/amazonsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
)

type MockClient struct {
	mock.Mock
	req      *http.Request
	response []byte
	err      error
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	m.req = req
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: 200, Body: r}, m.err
}

///

func TestAmazonSupport(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = nil

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../demo/test")
	options := websupport.Options{ResourceDirectory: resourcesDirectory}

	var session = sessions.NewCookieStore([]byte("super_secret"))
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(""))
		})
	}, options)
	router := server.Handler.(*mux.Router)
	router.Use(amazonsupport.NewAmazonSupport(mockClient, amazonsupport.AmazonCognitoConfiguration{}, session).Middleware)

	go websupport.Start(server, listener)
	healthsupport.WaitForHealthy(server)
	defer websupport.Stop(server)

	request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/?code=42", server.Addr), nil)
	response, _ := (&http.Client{}).Do(request)
	assert.Equal(t, "https://.auth..amazoncognito.com/oauth2/token?code=42&grant_type=authorization_code&redirect_uri=", mockClient.req.URL.String())

	body, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(body), "")

	mockClient.err = errors.New("oops.")
	erroneousRequest, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/?code=42", server.Addr), nil)
	erroneousResponse, _ := (&http.Client{}).Do(erroneousRequest)

	erroneousBody, _ := io.ReadAll(erroneousResponse.Body)
	assert.Contains(t, string(erroneousBody), "")
}
