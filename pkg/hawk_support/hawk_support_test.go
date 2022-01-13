package hawk_support_test

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/hawk_support"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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

func setup(key string) func(t *testing.T) {
	get := hawk_support.HawkMiddleware(secureGet, hawk_support.NewCredentialStore(key), "localhost:8883")
	post := hawk_support.HawkMiddleware(securePost, hawk_support.NewCredentialStore(key), "localhost:8883")
	server := web_support.Create("localhost:8883", func(router *mux.Router) {
		router.HandleFunc("/secure", get).Methods("GET")
		router.HandleFunc("/secure", post).Methods("POST")
	}, web_support.Options{})

	go web_support.Start(server)
	web_support.WaitForHealthy(server)

	return func(t *testing.T) {
		defer web_support.Stop(server)
	}
}

func TestGet(t *testing.T) {
	key := getKey()
	teardownTestCase := setup(key)
	defer teardownTestCase(t)

	resp, _ := hawk_support.HawkGet(&http.Client{}, "anId", key, "http://localhost:8883/secure")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "success", string(b))
}

func TestGet_fails(t *testing.T) {
	teardownTestCase := setup("")
	defer teardownTestCase(t)

	resp, _ := http.Get("http://localhost:8883/secure")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPost(t *testing.T) {
	key := getKey()
	teardownTestCase := setup(key)
	defer teardownTestCase(t)

	reader := strings.NewReader("aBody")
	resp, _ := hawk_support.HawkPost(&http.Client{}, "anId", key, "http://localhost:8883/secure", reader)
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
