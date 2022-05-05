package orchestrator_test

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net"
	"net/http"
	"testing"
)

type HandlerSuite struct {
	suite.Suite
	db      *sql.DB
	server  *http.Server
	key     string
	gateway orchestrator.IntegrationsDataGateway
}

func TestIntegrationsHandler(t *testing.T) {
	suite.Run(t, &HandlerSuite{})
}

func (s *HandlerSuite) SetupTest() {
	s.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	s.gateway = orchestrator.IntegrationsDataGateway{DB: s.db}
	_, _ = s.db.Exec("delete from integrations;")

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	s.key = hex.EncodeToString(hash[:])

	handlers, _ := orchestrator.LoadHandlers(s.db, hawksupport.NewCredentialStore(s.key), addr, map[string]orchestrator.Provider{})
	s.server = websupport.Create(addr, handlers, websupport.Options{})

	go websupport.Start(s.server, listener)
	healthsupport.WaitForHealthy(s.server)
}

func (s *HandlerSuite) TearDownTest() {
	_ = s.db.Close()
	websupport.Stop(s.server)
}

func (s *HandlerSuite) TestList() {
	_, _ = s.gateway.Create("aName", "google_cloud", []byte("aKey"))

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/integrations", s.server.Addr))
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var jsonResponse orchestrator.Integrations
	_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)

	integration := jsonResponse.Integrations[0]
	assert.Equal(s.T(), "aName", integration.Name)
	assert.Equal(s.T(), "google_cloud", integration.Provider)
	assert.Equal(s.T(), []byte("aKey"), integration.Key)
}

func (s *HandlerSuite) TestCreate() {
	integration := orchestrator.Integration{Name: "aName", Provider: "google_cloud", Key: []byte("aKey")}
	marshal, _ := json.Marshal(integration)
	_, _ = hawksupport.HawkPost(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/integrations", s.server.Addr), bytes.NewReader(marshal))

	records, _ := s.gateway.Find()
	assert.Equal(s.T(), 1, len(records))

	record := records[0]
	assert.Equal(s.T(), "aName", record.Name)
	assert.Equal(s.T(), "google_cloud", record.Provider)
	assert.Equal(s.T(), []byte("aKey"), record.Key)
}

func (s *HandlerSuite) TestDelete() {
	id, _ := s.gateway.Create("aName", "google_cloud", []byte("aKey"))

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/integrations/%s", s.server.Addr, id))
	assert.Equal(s.T(), resp.StatusCode, http.StatusOK)

	records, _ := s.gateway.Find()
	assert.Equal(s.T(), 0, len(records))
}

func (s *HandlerSuite) TestDelete_withUnknownID() {
	_, _ = s.gateway.Create("aName", "google_cloud", []byte("aKey"))

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/integrations/%s", s.server.Addr, "0000"))
	assert.Equal(s.T(), resp.StatusCode, http.StatusInternalServerError)
}
