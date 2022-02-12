package amazonsupport_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
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

type MockClaimsParser struct {
	Err error
}

func (m MockClaimsParser) ParseWithClaims(tokenString string, region string, claims jwt.Claims) (*jwt.Token, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	c := claims.(*amazonsupport.AmazonCognitoClaims)
	c.Email = "example@amazon.com"
	return nil, nil
}

///

func TestAmazonSupport(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = nil
	mockParser := MockClaimsParser{}

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../demo/test")
	options := websupport.Options{ResourceDirectory: resourcesDirectory}

	var session = sessions.NewCookieStore([]byte("super_secret"))
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			session, _ := session.Get(r, "session")
			principal := session.Values["principal"].([]string)
			_, _ = w.Write([]byte(principal[0]))
		})
	}, options)
	router := server.Handler.(*mux.Router)
	router.Use(amazonsupport.NewAmazonSupport(mockClient, amazonsupport.AmazonCognitoConfiguration{}, mockParser, session).Middleware)

	go websupport.Start(server, listener)
	healthsupport.WaitForHealthy(server)
	defer websupport.Stop(server)

	claims := &jwt.StandardClaims{ExpiresAt: 300, Issuer: "https://cognito", Id: "anId"}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = "billy"
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signedString, _ := token.SignedString(key)

	request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/", server.Addr), nil)
	request.Header["X-Amzn-Oidc-Data"] = []string{signedString}
	response, _ := (&http.Client{}).Do(request)

	body, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(body), "example@amazon.com")

	mockParser.Err = errors.New("oops")
	_, _ = (&http.Client{}).Do(request)
	assert.Contains(t, string(body), "")
}
