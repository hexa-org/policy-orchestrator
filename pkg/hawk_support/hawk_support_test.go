package hawk_support_test

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"hexa/pkg/hawk_support"
	"hexa/pkg/web_support"
	"io/ioutil"
	"net/http"
	"testing"
)

func secure(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("success"))
}

func setup(key string) func(t *testing.T) {
	server := web_support.Create("localhost:8883", func(x *mux.Router) {
		x.HandleFunc("/secure",
			hawk_support.HawkMiddleware(secure, hawk_support.NewCredentialStore(key), "localhost:8883"),
		).Methods("GET")
	}, web_support.Options{})

	go web_support.Start(server)
	web_support.WaitForHealthy(server)

	return func(t *testing.T) {
		defer web_support.Stop(server)
	}
}

func TestNotSecure(t *testing.T) {
	teardownTestCase := setup("")
	defer teardownTestCase(t)

	resp, _ := http.Get("http://localhost:8883/secure")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSecure(t *testing.T) {
	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	teardownTestCase := setup(key)
	defer teardownTestCase(t)

	resp, _ := hawk_support.HawkGet(&http.Client{}, "anId", key, "http://localhost:8883/secure")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "success", string(b))
}
