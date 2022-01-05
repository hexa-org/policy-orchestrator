package orchestrator_test

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/web_support"
	"log"
	"net/http"
	"testing"
)


func TestOrchestratorHandlers(t *testing.T) {
	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	store := hawk_support.NewCredentialStore(key)
	handlers := orchestrator.LoadHandlers(store, "localhost:8883")
	server := web_support.Create("localhost:8883", handlers, web_support.Options{})
	go web_support.Start(server)
	web_support.WaitForHealthy(server)

	resp, err := http.Get("http://localhost:8883/health")
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	web_support.Stop(server)
}
