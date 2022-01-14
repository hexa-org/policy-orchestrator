package hawk_support_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/hawk_support"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
)

func secureGet(w http.ResponseWriter, r *http.Request) {
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
	get := hawk_support.HawkMiddleware(secureGet, hawk_support.NewCredentialStore(key), listener.Addr().String())
	post := hawk_support.HawkMiddleware(securePost, hawk_support.NewCredentialStore(key), listener.Addr().String())
	server := web_support.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/secure", get).Methods("GET")
		router.HandleFunc("/secure", post).Methods("POST")
	}, web_support.Options{})

	go web_support.Start(server, listener)
	web_support.WaitForHealthy(server)

	return server, func(t *testing.T) {
		defer web_support.Stop(server)
	}
}

func TestGet(t *testing.T) {
	key := getKey()
	app, teardownTestCase := setup(key)
	defer teardownTestCase(t)

	resp, _ := hawk_support.HawkGet(&http.Client{}, "anId", key, fmt.Sprintf("http://%s/secure", app.Addr))
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

	_, err := hawk_support.HawkGet(&http.Client{}, "anId", key, "httq://localhost")
	assert.Error(t, err)
}

func TestPost(t *testing.T) {
	key := getKey()
	app, teardownTestCase := setup(key)
	defer teardownTestCase(t)

	reader := strings.NewReader("aBody")
	resp, _ := hawk_support.HawkPost(&http.Client{}, "anId", key, fmt.Sprintf("http://%s/secure", app.Addr), reader)
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
