package orchestrator_test

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"hexa/pkg/database_support"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/web_support"
	"hexa/pkg/workflow_support"
	"net/http"
)

type SuiteFields struct {
	DB        *sql.DB
	Server    *http.Server
	Scheduler *workflow_support.WorkScheduler
	Key       string
	Gateway   orchestrator.IntegrationsDataGateway
}

func (fields *SuiteFields) Setup() {
	fields.DB, _ = database_support.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	fields.Gateway = orchestrator.IntegrationsDataGateway{DB: fields.DB}
	_, _ = fields.DB.Exec("delete from applications;")
	_, _ = fields.DB.Exec("delete from integrations;")

	hash := sha256.Sum256([]byte("aKey"))
	fields.Key = hex.EncodeToString(hash[:])

	handlers, scheduler := orchestrator.LoadHandlers(hawk_support.NewCredentialStore(fields.Key), "localhost:8883", fields.DB)
	fields.Scheduler = scheduler
	fields.Server = web_support.Create("localhost:8883", handlers, web_support.Options{})
}
