package orchestrator_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestOrchestratorHandlers(t *testing.T) {
	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")

	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	store := hawksupport.NewCredentialStore(key)

	listener, _ := net.Listen("tcp", "localhost:0")

	providers := make(map[string]policyprovider.Provider)
	providers["noop"] = &orchestrator_test.NoopProvider{}

	handlers, _ := orchestrator.LoadHandlers(db, store, listener.Addr().String(), providers)
	server := websupport.Create(listener.Addr().String(), handlers, websupport.Options{})

	go websupport.Start(server, listener)

	healthsupport.WaitForHealthy(server)

	resp, err := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	websupport.Stop(server)
}
