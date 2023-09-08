package orchestrator_test

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

type orchestrationHandlerData struct {
	db        *sql.DB
	server    *http.Server
	key       string
	providers map[string]orchestrator.Provider

	fromApp        string
	toApp          string
	toAppDifferent string
}

func (data *orchestrationHandlerData) SetUp() {
	data.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = data.db.Exec(`
delete from applications;
delete from integrations;
insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'aName', 'noop', 'aKey');
insert into applications (id, integration_id, object_id, name, description) values ('6409776a-367a-483a-a194-5ccf9c4ff210', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anObjectId', 'aName', 'aDescription');
insert into applications (id, integration_id, object_id, name, description) values ('6409776a-367a-483a-a194-5ccf9c4ff211', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anotherObjectId', 'anotherName', 'anotherDescription');

insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e8', 'anotherName', 'azure', 'aKey');
insert into applications (id, integration_id, object_id, name, description) values ('6409776a-367a-483a-a194-5ccf9c4ff212', '50e00619-9f15-4e85-a7e9-f26d87ea12e8', 'andAnotherObjectId', 'andAnotherName', 'andAnotherDescription');
`)
	data.fromApp = "6409776a-367a-483a-a194-5ccf9c4ff210"
	data.toApp = "6409776a-367a-483a-a194-5ccf9c4ff211"
	data.toAppDifferent = "6409776a-367a-483a-a194-5ccf9c4ff212"

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	data.key = hex.EncodeToString(hash[:])

	data.providers = make(map[string]orchestrator.Provider)
	data.providers["noop"] = &orchestrator_test.NoopProvider{}
	data.providers["azure"] = microsoftazure.NewAzureProvider()
	handlers, _ := orchestrator.LoadHandlers(data.db, hawksupport.NewCredentialStore(data.key), addr, data.providers)
	data.server = websupport.Create(addr, handlers, websupport.Options{})
	go websupport.Start(data.server, listener)
	healthsupport.WaitForHealthy(data.server)
}

func (data *orchestrationHandlerData) TearDown() {
	_ = data.db.Close()
	websupport.Stop(data.server)
}

func TestOrchestration(t *testing.T) {
	testsupport.WithSetUp(&orchestrationHandlerData{}, func(data *orchestrationHandlerData) {
		url := fmt.Sprintf("http://%s/orchestration", data.server.Addr)
		marshal, _ := json.Marshal(orchestrator.Orchestration{From: data.fromApp, To: data.toApp})

		resp, _ := hawksupport.HawkPost(&http.Client{}, "anId", data.key, url, bytes.NewReader(marshal))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestOrchestration_failsAcrossProviders(t *testing.T) {
	testsupport.WithSetUp(&orchestrationHandlerData{}, func(data *orchestrationHandlerData) {
		url := fmt.Sprintf("http://%s/orchestration", data.server.Addr)
		marshal, _ := json.Marshal(orchestrator.Orchestration{From: data.fromApp, To: data.toAppDifferent})

		resp, err := hawksupport.HawkPost(&http.Client{}, "anId", data.key, url, bytes.NewReader(marshal))
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
