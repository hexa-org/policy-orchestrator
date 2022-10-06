package orchestrator_test

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

type applicationsHandlerData struct {
	db                *sql.DB
	server            *http.Server
	key               string
	providers         map[string]orchestrator.Provider
	applicationTestId string
}

func (data *applicationsHandlerData) SetUp() {
	data.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = data.db.Exec(`
delete from applications;
delete from integrations;
insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'aName', 'noop', 'aKey');
insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff210', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anObjectId', 'aName', 'aDescription', 'aService');
`)
	data.applicationTestId = "6409776a-367a-483a-a194-5ccf9c4ff210"

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	data.key = hex.EncodeToString(hash[:])

	data.providers = make(map[string]orchestrator.Provider)
	data.providers["noop"] = &orchestrator_test.NoopProvider{}
	handlers, _ := orchestrator.LoadHandlers(data.db, hawksupport.NewCredentialStore(data.key), addr, data.providers)
	data.server = websupport.Create(addr, handlers, websupport.Options{})
	go websupport.Start(data.server, listener)
	healthsupport.WaitForHealthy(data.server)
}

func (data *applicationsHandlerData) TearDown() {
	_ = data.db.Close()
	websupport.Stop(data.server)
}

func TestListApps(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		url := fmt.Sprintf("http://%s/applications", data.server.Addr)

		resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", data.key, url)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apps orchestrator.Applications
		_ = json.NewDecoder(resp.Body).Decode(&apps)
		assert.Equal(t, 1, len(apps.Applications))

		application := apps.Applications[0]
		assert.Equal(t, "anObjectId", application.ObjectId)
		assert.Equal(t, "aName", application.Name)
		assert.Equal(t, "noop", application.ProviderName)
		assert.Equal(t, "aDescription", application.Description)
		assert.Equal(t, "aService", application.Service)
	})
}

func TestListApps_withErroneousDatabase(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		_ = data.db.Close()

		url := fmt.Sprintf("http://%s/applications", data.server.Addr)

		resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", data.key, url)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestShowApps(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		url := fmt.Sprintf("http://%s/applications/%s", data.server.Addr, data.applicationTestId)

		resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", data.key, url)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var app orchestrator.Application
		_ = json.NewDecoder(resp.Body).Decode(&app)
		assert.Equal(t, "anObjectId", app.ObjectId)
		assert.Equal(t, "aName", app.Name)
		assert.Equal(t, "aDescription", app.Description)
	})
}

func TestShowApps_withUnknownID(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		url := fmt.Sprintf("http://%s/applications/oops", data.server.Addr)

		resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", data.key, url)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetPolicies(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		url := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", data.key, url)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var policies orchestrator.Policies
		_ = json.NewDecoder(resp.Body).Decode(&policies)
		assert.Equal(t, 2, len(policies.Policies))

		policy := policies.Policies[0]
		assert.Equal(t, "anAction", policy.Actions[0].ActionUri)
		assert.Equal(t, "aVersion", policy.Meta.Version)
		assert.Equal(t, []string{"aUser"}, policy.Subject.Members)
		assert.Equal(t, "anId", policy.Object.ResourceID)
	})
}

func TestGetPolicies_withDatabaseError(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		_ = data.db.Close()

		url := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", data.key, url)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetPolicies_withFailedRequest(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		discovery := orchestrator_test.NoopProvider{}
		discovery.Err = errors.New("oops")
		data.providers["noop"] = &discovery

		url := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", data.key, url)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestSetPolicies(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		var buf bytes.Buffer
		policy := orchestrator.Policy{
			Meta:    orchestrator.Meta{Version: "v0.5"},
			Actions: []orchestrator.Action{{"anAction"}},
			Subject: orchestrator.Subject{Members: []string{"anEmail", "anotherEmail"}},
			Object: orchestrator.Object{
				ResourceID: "aResourceId",
			},
		}
		_ = json.NewEncoder(&buf).Encode(orchestrator.Policies{Policies: []orchestrator.Policy{policy}})

		url := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", data.key, url, bytes.NewReader(buf.Bytes()))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestSetPolicies_withDatabaseError(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		_ = data.db.Close()

		url := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", data.key, url, bytes.NewReader([]byte("")))
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestSetPolicies_withErroneousProvider(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		noopProvider := orchestrator_test.NoopProvider{}
		noopProvider.Err = errors.New("oops")
		data.providers["noop"] = &noopProvider

		var buf bytes.Buffer
		policy := orchestrator.Policy{Meta: orchestrator.Meta{Version: "v0.5"}, Actions: []orchestrator.Action{{"anAction"}}, Subject: orchestrator.Subject{Members: []string{"anEmail", "anotherEmail"}}, Object: orchestrator.Object{ResourceID: "aResourceId"}}
		_ = json.NewEncoder(&buf).Encode(policy)

		url := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", data.key, url, bytes.NewReader(buf.Bytes()))
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestSetPolicies_withMissingJson(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		url := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, "anId")

		resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", data.key, url, nil)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
