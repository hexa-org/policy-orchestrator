package orchestrator_test

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"net/http"
)

type SuiteFields struct {
	DB        *sql.DB
	Server    *http.Server
	Scheduler *workflowsupport.WorkScheduler
	Key       string
	Gateway   orchestrator.IntegrationsDataGateway
	Providers map[string]provider.Provider
}

func (fields *SuiteFields) Setup(providers map[string]provider.Provider, addr string) {
	fields.DB, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	fields.Gateway = orchestrator.IntegrationsDataGateway{DB: fields.DB}
	_, _ = fields.DB.Exec("delete from applications;")
	_, _ = fields.DB.Exec("delete from integrations;")

	hash := sha256.Sum256([]byte("aKey"))
	fields.Key = hex.EncodeToString(hash[:])
	fields.Providers = providers
	handlers, scheduler := orchestrator.LoadHandlers(fields.DB, hawksupport.NewCredentialStore(fields.Key), addr, fields.Providers)
	fields.Scheduler = scheduler
	fields.Server = websupport.Create(addr, handlers, websupport.Options{})
}
