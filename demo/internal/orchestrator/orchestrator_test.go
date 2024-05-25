package orchestrator_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestOrchestratorHandlers(t *testing.T) {

	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	store := hawksupport.NewCredentialStore(key)

	listener, _ := net.Listen("tcp", "localhost:0")

	tempDir, err := os.MkdirTemp("", "hexa-orchestrator-*")
	assert.NoError(t, err, "Error creating temp dir")

	testConfigPath := filepath.Join(tempDir, ".hexa", "config.json")

	_ = os.Setenv(dataConfigGateway.EnvIntegrationConfigFile, testConfigPath)

	config, err := dataConfigGateway.NewIntegrationConfigData()
	assert.NoError(t, err, "Error initializing config")
	handlers := orchestrator.LoadHandlers(config, store, listener.Addr().String(), nil)
	server := websupport.Create(listener.Addr().String(), handlers, websupport.Options{})

	go websupport.Start(server, listener)

	healthsupport.WaitForHealthy(server)

	resp, err := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	websupport.Stop(server)

	_ = os.RemoveAll(tempDir)
}
