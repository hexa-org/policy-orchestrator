package hawksupport_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
)

func secureGet(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func securePost(w http.ResponseWriter, r *http.Request) {
	all, _ := ioutil.ReadAll(r.Body)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(all)
}

func setup(key string) (*http.Server, func(t *testing.T)) {
	listener, _ := net.Listen("tcp", "localhost:0")
	get := hawksupport.HawkMiddleware(secureGet, hawksupport.NewCredentialStore(key), listener.Addr().String())
	post := hawksupport.HawkMiddleware(securePost, hawksupport.NewCredentialStore(key), listener.Addr().String())
	server := websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/secure", get).Methods("GET")
		router.HandleFunc("/secure", post).Methods("POST")
	}, websupport.Options{})

	go websupport.Start(server, listener)
	healthsupport.WaitForHealthy(server)

	return server, func(t *testing.T) {
		defer websupport.Stop(server)
	}
}

func TestGet(t *testing.T) {
	key := getKey()
	app, teardownTestCase := setup(key)
	defer teardownTestCase(t)

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", key, fmt.Sprintf("http://%s/secure", app.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "success", string(b))
}

func TestGet_fails(t *testing.T) {
	app, teardownTestCase := setup("")
	defer teardownTestCase(t)

	resp, _ := http.Get(fmt.Sprintf("http://%s/secure", app.Addr))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGet_bad_url(t *testing.T) {
	key := getKey()
	_, teardownTestCase := setup(key)
	defer teardownTestCase(t)

	_, err := hawksupport.HawkGet(&http.Client{}, "anId", key, "httq://localhost")
	assert.Error(t, err)
}

func TestGet_bad_hostname(t *testing.T) {
	key := getKey()
	app, teardownTestCase := setup(key)
	defer teardownTestCase(t)

	port := strings.Split(app.Addr, ":")[1]
	hostport := "localhost:" + port
	resp, err := hawksupport.HawkGet(&http.Client{}, "anId", key, fmt.Sprintf("http://%s/secure", hostport))
	assert.Equal(t, err, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPost(t *testing.T) {
	key := getKey()
	app, teardownTestCase := setup(key)
	defer teardownTestCase(t)

	reader := strings.NewReader("aBody")
	resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", key, fmt.Sprintf("http://%s/secure", app.Addr), reader)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "aBody", string(b))
}

///

func getKey() string {
	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	return key
}
