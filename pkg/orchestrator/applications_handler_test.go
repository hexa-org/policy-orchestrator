package orchestrator_test

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"net"
	"net/http"
	"testing"
)

type ApplicationsHandlerSuite struct {
	suite.Suite
	db         *sql.DB
	server     *http.Server
	scheduler  *workflowsupport.WorkScheduler
	key        string
	gateway    orchestrator.IntegrationsDataGateway
	appGateway orchestrator.ApplicationsDataGateway
	providers  map[string]provider.Provider
}

func TestApplicationsHandler(t *testing.T) {
	suite.Run(t, &ApplicationsHandlerSuite{})
}

func (s *ApplicationsHandlerSuite) SetupTest() {
	s.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	s.gateway = orchestrator.IntegrationsDataGateway{DB: s.db}

	// todo - move below to scenario style
	_, _ = s.db.Exec("delete from applications;")
	_, _ = s.db.Exec("delete from integrations;")

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	s.key = hex.EncodeToString(hash[:])

	s.providers = make(map[string]provider.Provider)

	handlers, scheduler := orchestrator.LoadHandlers(s.db, hawksupport.NewCredentialStore(s.key), addr, s.providers)
	s.scheduler = scheduler
	s.server = websupport.Create(addr, handlers, websupport.Options{})

	go websupport.Start(s.server, listener)
	websupport.WaitForHealthy(s.server)
}

func (s *ApplicationsHandlerSuite) TearDownTest() {
	_ = s.db.Close()
	websupport.Stop(s.server)
}

func (s *ApplicationsHandlerSuite) TestList() {
	var integrationTestId string
	_ = s.db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "aProvider", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = s.db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	s.providers["google cloud"] = &orchestrator_test.NoopProvider{}

	resp, err := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications", s.server.Addr))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var apps orchestrator.Applications
	_ = json.NewDecoder(resp.Body).Decode(&apps)
	assert.Equal(s.T(), 1, len(apps.Applications))
	assert.Equal(s.T(), "anObjectId", apps.Applications[0].ObjectId)
	assert.Equal(s.T(), "aName", apps.Applications[0].Name)
	assert.Equal(s.T(), "aDescription", apps.Applications[0].Description)
}

func (s *ApplicationsHandlerSuite) TestShow() {
	var integrationTestId string
	_ = s.db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "aProvider", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = s.db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	s.providers["google cloud"] = &orchestrator_test.NoopProvider{}

	resp, err := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications/%s", s.server.Addr, applicationTestId))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var app orchestrator.Application
	_ = json.NewDecoder(resp.Body).Decode(&app)
	assert.Equal(s.T(), "anObjectId", app.ObjectId)
	assert.Equal(s.T(), "aName", app.Name)
	assert.Equal(s.T(), "aDescription", app.Description)
}

func (s *ApplicationsHandlerSuite) TestShow_withUnknownID() {
	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications/oops", s.server.Addr))
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestGetPolicies() {
	var integrationTestId string
	_ = s.db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "google cloud", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = s.db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	s.providers["google cloud"] = &orchestrator_test.NoopProvider{}

	resp, err := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, applicationTestId))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var p []orchestrator.Policy
	_ = json.NewDecoder(resp.Body).Decode(&p)
	assert.Equal(s.T(), 2, len(p))
	assert.Equal(s.T(), "anAction", p[0].Action)
	assert.Equal(s.T(), "aVersion", p[0].Version)
	assert.Equal(s.T(), []string{"aUser"}, p[0].Subject.AuthenticatedUsers)
	assert.Equal(s.T(), []string{"/"}, p[0].Object.Resources)
}

func (s *ApplicationsHandlerSuite) TestGetPolicies_withRequestFails() {
	var integrationTestId string
	_ = s.db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "google cloud", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = s.db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	discovery := orchestrator_test.NoopProvider{}
	discovery.Err = errors.New("oops")
	s.providers["google cloud"] = &discovery

	resp, err := hawksupport.HawkGet(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, applicationTestId))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestSetPolicies() {
	var integrationTestId string
	_ = s.db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "google cloud", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = s.db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	s.providers["google cloud"] = &orchestrator_test.NoopProvider{}

	var buf bytes.Buffer
	policy := orchestrator.Policy{Version: "v0.1", Action: "anAction", Subject: orchestrator.Subject{AuthenticatedUsers: []string{"anEmail", "anotherEmail"}}, Object: orchestrator.Object{Resources: []string{"/"}}}
	_ = json.NewEncoder(&buf).Encode([]orchestrator.Policy{policy})

	resp, err := hawksupport.HawkPost(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, applicationTestId), bytes.NewReader(buf.Bytes()))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestSetPolicies_withErroneousProvider() {
	var integrationTestId string
	_ = s.db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "google cloud", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = s.db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	noopProvider := orchestrator_test.NoopProvider{}
	noopProvider.Err = errors.New("oops")
	s.providers["google cloud"] = &noopProvider

	var buf bytes.Buffer
	policy := orchestrator.Policy{Version: "v0.1", Action: "anAction", Subject: orchestrator.Subject{[]string{"anEmail", "anotherEmail"}}, Object: orchestrator.Object{Resources: []string{"/"}}}
	_ = json.NewEncoder(&buf).Encode(policy)

	resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, applicationTestId), bytes.NewReader(buf.Bytes()))
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestSetPolicies_withMissingJson() {
	resp, err := hawksupport.HawkPost(&http.Client{}, "anId", s.key, fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, "anId"), nil)
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}
