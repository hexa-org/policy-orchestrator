package main

import (
	"database/sql"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/googlecloud"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/openpolicyagent"
	log "golang.org/x/exp/slog"
	"net"
	"net/http"
	"os"

	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
)

type DatabaseHealthCheck struct {
	Db *sql.DB
}

func (d DatabaseHealthCheck) Name() string {
	return "database"
}

func (d DatabaseHealthCheck) Check() bool {
	err := d.Db.Ping()
	if err != nil {
		return false
	}
	return true
}

type ServerHealthCheck struct {
}

func (s ServerHealthCheck) Name() string {
	return "server"
}

func (s ServerHealthCheck) Check() bool {
	return true
}

func App(key string, addr string, hostPort string, dbUrl string) (*http.Server, *workflowsupport.WorkScheduler) {
	db, _ := databasesupport.Open(dbUrl)
	store := hawksupport.NewCredentialStore(key)
	providers := make(map[string]orchestrator.Provider)
	providers["google_cloud"] = &googlecloud.GoogleProvider{}
	providers["azure"] = microsoftazure.NewAzureApimProvider()
	//providers["azure_apim"] = microsoftazure.NewAzureApimProvider()
	providers["amazon"] = &amazonwebservices.AmazonProvider{}
	providers["open_policy_agent"] = &openpolicyagent.OpaProvider{}
	handlers, scheduler := orchestrator.LoadHandlers(db, store, hostPort, providers)
	return websupport.Create(addr, handlers, websupport.Options{
		HealthChecks: []healthsupport.HealthCheck{
			ServerHealthCheck{},
			DatabaseHealthCheck{db},
		},
	}), scheduler
}

func newApp(addr string) (*http.Server, net.Listener, *workflowsupport.WorkScheduler) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Info("Found server address %v", addr)

	if found := os.Getenv("HOST"); found != "" {
		_, port, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", found, port)
	}
	log.Info("Found server host %v", addr)

	dbUrl := os.Getenv("POSTGRESQL_URL")
	key := os.Getenv("ORCHESTRATOR_KEY")
	hostPort := os.Getenv("ORCHESTRATOR_HOSTPORT")
	listener, _ := net.Listen("tcp", addr)
	app, scheduler := App(key, listener.Addr().String(), hostPort, dbUrl)

	if certFile := os.Getenv("SERVER_CERT_PATH"); certFile != "" {
		keyFile := os.Getenv("SERVER_KEY_PATH")
		websupport.WithTransportLayerSecurity(certFile, keyFile, app)
	}

	return app, listener, scheduler
}

func main() {
	app, listener, scheduler := newApp("0.0.0.0:8885")
	scheduler.Start()
	websupport.Start(app, listener)
}
