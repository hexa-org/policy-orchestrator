package orchestrator_test

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"hexa/pkg/database_support"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/web_support"
	"io"
	"log"
	"net/http"
	"testing"
)

func setup(key string) func(t *testing.T) {
	db, _ := database_support.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")

	handlers, _ := orchestrator.LoadHandlers(hawk_support.NewCredentialStore(key), "localhost:8883", db)
	server := web_support.Create("localhost:8883", handlers, web_support.Options{})

	go web_support.Start(server)
	web_support.WaitForHealthy(server)

	return func(t *testing.T) {
		defer web_support.Stop(server)
	}
}

func TestApplications(t *testing.T) {
	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	teardownTestCase := setup(key)
	defer teardownTestCase(t)

	resp, err := hawk_support.HawkGet(&http.Client{}, "anId", key, "http://localhost:8883/applications")
	if err != nil {
		log.Fatalln(err)
	}
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"applications\":[{\"name\":\"anApp\"}]}", string(body))
}
