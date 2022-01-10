package orchestrator_test

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"hexa/pkg/database_support"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/web_support"
	"hexa/pkg/workflow_support"
	"net/http"
	"testing"
)

type HandlerSuite struct {
	suite.Suite
	db      *sql.DB
	server  *http.Server
	scheduler *workflow_support.WorkScheduler
	key     string
	gateway orchestrator.IntegrationsDataGateway
}

func TestIntegrationsHandler(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

func (suite *HandlerSuite) SetupTest() {
	suite.db, _ = database_support.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	suite.gateway = orchestrator.IntegrationsDataGateway{DB: suite.db}
	_, _ = suite.db.Exec("delete from integrations;")

	hash := sha256.Sum256([]byte("aKey"))
	suite.key = hex.EncodeToString(hash[:])

	handlers, scheduler := orchestrator.LoadHandlers(hawk_support.NewCredentialStore(suite.key), "localhost:8883", suite.db)
	suite.scheduler = scheduler
	suite.server = web_support.Create("localhost:8883", handlers, web_support.Options{})

	go web_support.Start(suite.server)
	web_support.WaitForHealthy(suite.server)
}

func (suite *HandlerSuite) TearDownTest() {
	_ = suite.db.Close()
	web_support.Stop(suite.server)
}

func (suite *HandlerSuite) TestList() {
	_, _ = suite.gateway.Create("aName", "google cloud", []byte("aKey"))

	resp, _ := hawk_support.HawkGet(&http.Client{}, "anId", suite.key, "http://localhost:8883/integrations")
	var jsonResponse orchestrator.Integrations
	_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "aName", jsonResponse.Integrations[0].Name)
	assert.Equal(suite.T(), "google cloud", jsonResponse.Integrations[0].Provider)
	assert.Equal(suite.T(), []byte("aKey"), jsonResponse.Integrations[0].Key)
}

func (suite *HandlerSuite) TestCreate() {
	integration := orchestrator.Integration{Name: "aName", Provider: "google cloud", Key: []byte("aKey")}
	marshal, _ := json.Marshal(integration)
	_, _ = hawk_support.HawkPost(&http.Client{}, "anId", suite.key, "http://localhost:8883/integrations", bytes.NewReader(marshal))

	all, _ := suite.gateway.Find()
	assert.Equal(suite.T(), 1, len(all))
	assert.Equal(suite.T(), "aName", all[0].Name)
	assert.Equal(suite.T(), "google cloud", all[0].Provider)
	assert.Equal(suite.T(), []byte("aKey"), all[0].Key)
}

func (suite *HandlerSuite) TestDelete() {
	id, _ := suite.gateway.Create("aName", "google cloud", []byte("aKey"))
	resp, _ := hawk_support.HawkGet(&http.Client{}, "anId", suite.key, fmt.Sprintf("http://localhost:8883/integrations/%s", id))
	assert.Equal(suite.T(), resp.StatusCode, http.StatusOK)

	all, _ := suite.gateway.Find()
	assert.Equal(suite.T(), 0, len(all))
}
