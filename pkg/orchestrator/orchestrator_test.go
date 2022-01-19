package orchestrator_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	orchestrator_test "github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"net/http"
	"testing"
)

func TestOrchestratorHandlers(t *testing.T) {
	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")

	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	store := hawksupport.NewCredentialStore(key)

	listener, _ := net.Listen("tcp", "localhost:0")

	providers := make(map[string]provider.Provider)
	providers["google cloud"] = &orchestrator_test.NoopDiscovery{}

	handlers, _ := orchestrator.LoadHandlers(db, store, listener.Addr().String(), providers)
	server := websupport.Create(listener.Addr().String(), handlers, websupport.Options{})

	go websupport.Start(server, listener)

	websupport.WaitForHealthy(server)

	resp, err := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	websupport.Stop(server)
}
