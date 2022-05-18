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
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net"
	"net/http"
	"testing"
)

type ApplicationsHandlerSuite struct {
	suite.Suite
	db                *sql.DB
	server            *http.Server
	scheduler         *workflowsupport.WorkScheduler
	key               string
	providers         map[string]orchestrator.Provider
	applicationTestId string
}

func TestApplicationsHandler(t *testing.T) {
	suite.Run(t, &ApplicationsHandlerSuite{})
}

func (s *ApplicationsHandlerSuite) SetupTest() {
	s.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = s.db.Exec(`
delete from applications;
delete from integrations;
insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'aName', 'google_cloud', 'aKey');
insert into applications (id, integration_id, object_id, name, description) values ('6409776a-367a-483a-a194-5ccf9c4ff210', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anObjectId', 'aName', 'aDescription');
`)
	s.applicationTestId = "6409776a-367a-483a-a194-5ccf9c4ff210"

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	s.key = hex.EncodeToString(hash[:])

	s.providers = make(map[string]orchestrator.Provider)
	s.providers["google_cloud"] = &orchestrator_test.NoopProvider{}

	handlers, scheduler := orchestrator.LoadHandlers(s.db, hawksupport.NewCredentialStore(s.key), addr, s.providers)
	s.scheduler = scheduler
	s.server = websupport.Create(addr, handlers, websupport.Options{})

	go websupport.Start(s.server, listener)
	healthsupport.WaitForHealthy(s.server)
}

func (s *ApplicationsHandlerSuite) TearDownTest() {
	_ = s.db.Close()
	websupport.Stop(s.server)
}

func (s *ApplicationsHandlerSuite) TestList() {
	url := fmt.Sprintf("http://%s/applications", s.server.Addr)

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, url)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var apps orchestrator.Applications
	_ = json.NewDecoder(resp.Body).Decode(&apps)
	assert.Equal(s.T(), 1, len(apps.Applications))

	application := apps.Applications[0]
	assert.Equal(s.T(), "anObjectId", application.ObjectId)
	assert.Equal(s.T(), "aName", application.Name)
	assert.Equal(s.T(), "aDescription", application.Description)
}

func (s *ApplicationsHandlerSuite) TestShow() {
	url := fmt.Sprintf("http://%s/applications/%s", s.server.Addr, s.applicationTestId)

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, url)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var app orchestrator.Application
	_ = json.NewDecoder(resp.Body).Decode(&app)
	assert.Equal(s.T(), "anObjectId", app.ObjectId)
	assert.Equal(s.T(), "aName", app.Name)
	assert.Equal(s.T(), "aDescription", app.Description)
}

func (s *ApplicationsHandlerSuite) TestShow_withUnknownID() {
	url := fmt.Sprintf("http://%s/applications/oops", s.server.Addr)

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, url)
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestGetPolicies() {
	url := fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, s.applicationTestId)

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, url)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var policies orchestrator.Policies
	_ = json.NewDecoder(resp.Body).Decode(&policies)
	assert.Equal(s.T(), 2, len(policies.Policies))

	policy := policies.Policies[0]
	assert.Equal(s.T(), "anAction", policy.Action)
	assert.Equal(s.T(), "aVersion", policy.Version)
	assert.Equal(s.T(), []string{"aUser"}, policy.Subject.AuthenticatedUsers)
	assert.Equal(s.T(), []string{"/"}, policy.Object.Resources)
}

func (s *ApplicationsHandlerSuite) TestGetPolicies_withRequestFails() {
	discovery := orchestrator_test.NoopProvider{}
	discovery.Err = errors.New("oops")
	s.providers["google_cloud"] = &discovery

	url := fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, s.applicationTestId)

	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", s.key, url)
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestSetPolicies() {
	var buf bytes.Buffer
	policy := orchestrator.Policy{Version: "v0.1", Action: "anAction", Subject: orchestrator.Subject{AuthenticatedUsers: []string{"anEmail", "anotherEmail"}}, Object: orchestrator.Object{Resources: []string{"/"}}}
	_ = json.NewEncoder(&buf).Encode(orchestrator.Policies{Policies: []orchestrator.Policy{policy}})

	url := fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, s.applicationTestId)
	resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", s.key, url, bytes.NewReader(buf.Bytes()))
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestSetPolicies_withErroneousProvider() {
	noopProvider := orchestrator_test.NoopProvider{}
	noopProvider.Err = errors.New("oops")
	s.providers["google_cloud"] = &noopProvider

	var buf bytes.Buffer
	policy := orchestrator.Policy{Version: "v0.1", Action: "anAction", Subject: orchestrator.Subject{AuthenticatedUsers: []string{"anEmail", "anotherEmail"}}, Object: orchestrator.Object{Resources: []string{"/"}}}
	_ = json.NewEncoder(&buf).Encode(policy)

	url := fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, s.applicationTestId)

	resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", s.key, url, bytes.NewReader(buf.Bytes()))
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (s *ApplicationsHandlerSuite) TestSetPolicies_withMissingJson() {
	url := fmt.Sprintf("http://%s/applications/%s/policies", s.server.Addr, "anId")

	resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", s.key, url, nil)
	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)
}
