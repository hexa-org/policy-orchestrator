package orchestrator_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/database_support"
	"github.com/hexa-org/policy-orchestrator/pkg/hawk_support"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"net/http"
	"testing"
)

func TestOrchestratorHandlers(t *testing.T) {
	db, _ := database_support.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	store := hawk_support.NewCredentialStore(key)
	listener, _ := net.Listen("tcp", "localhost:0")
	handlers, _ := orchestrator.LoadHandlers(store, listener.Addr().String(), db)
	server := web_support.Create(listener.Addr().String(), handlers, web_support.Options{})
	go web_support.Start(server, listener)
	web_support.WaitForHealthy(server)

	resp, err := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	web_support.Stop(server)
}
