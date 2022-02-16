package orchestrator_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setUp() (orchestrator.IntegrationsDataGateway, orchestrator.ApplicationsDataGateway) {
	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = db.Exec("delete from integrations;")
	gateway := orchestrator.IntegrationsDataGateway{DB: db}
	appGateway := orchestrator.ApplicationsDataGateway{DB: db}
	return gateway, appGateway
}

func TestWorkflow(t *testing.T) {
	gateway, appGateway := setUp()
	_, _ = gateway.Create("aName", "noop", []byte("aKey"))

	noopProvider := orchestrator_test.NoopProvider{}
	providers := make(map[string]provider.Provider)
	providers["noop"] = &noopProvider
	worker := orchestrator.DiscoveryWorker{Providers: providers, Gateway: appGateway}
	finder := orchestrator.NewDiscoveryWorkFinder(gateway)
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)

	scheduler.Start()
	assert.True(t, <-finder.Results)
	scheduler.Stop()

	find, _ := appGateway.Find()
	assert.Equal(t, 3, len(find))
	assert.True(t, noopProvider.Discovered > 2)
}

func TestWorkflow_withEmptyResults(t *testing.T) {
	gateway, _ := setUp()

	worker := orchestrator.DiscoveryWorker{}
	finder := orchestrator.NewDiscoveryWorkFinder(gateway)
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)

	scheduler.Start()
	time.Sleep(time.Duration(50) * time.Millisecond)
	scheduler.Stop()

	_, resultReceived := <-finder.Results
	assert.False(t, resultReceived)
}

type ErroneousWorker struct {
}

func (n *ErroneousWorker) Run(interface{}) error {
	return errors.New("oops")
}

func TestWorkflow_erroneousFind(t *testing.T) {
	gateway, _ := setUp()
	_, _ = gateway.Create("aName", "google_cloud", []byte("aKey"))

	worker := ErroneousWorker{}
	finder := orchestrator.NewDiscoveryWorkFinder(gateway)
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)

	scheduler.Start()
	assert.False(t, <-finder.Results)
	scheduler.Stop()
}
