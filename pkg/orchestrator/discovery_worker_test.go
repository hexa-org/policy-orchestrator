package orchestrator_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"hexa/pkg/database_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/orchestrator/provider"
	"hexa/pkg/orchestrator/test"
	"hexa/pkg/workflow_support"
	"testing"
	"time"
)

func setUp() orchestrator.IntegrationsDataGateway {
	db, _ := database_support.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = db.Exec("delete from integrations;")
	gateway := orchestrator.IntegrationsDataGateway{DB: db}
	return gateway
}

func TestWorkflow(t *testing.T) {
	gateway := setUp()
	_, _ = gateway.Create("aName", "noop", []byte("aKey"))

	discovery := orchestrator_test.NoopDiscovery{}
	worker := orchestrator.DiscoveryWorker{Providers: []provider.Provider{&discovery}}
	finder := orchestrator.DiscoveryWorkFinder{Gateway: gateway}
	list := []workflow_support.Worker{&worker}
	scheduler := workflow_support.WorkScheduler{Finder: &finder, Workers: list, Delay: 50}
	scheduler.Start()
	for finder.Completed < 1 {
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
	scheduler.Stop()

	assert.Equal(t, 1, finder.Completed)
	assert.True(t, discovery.Discovered > 2)
}

func TestWorkflow_empty(t *testing.T) {
	gateway := setUp()

	worker := orchestrator.DiscoveryWorker{}
	finder := orchestrator.DiscoveryWorkFinder{Gateway: gateway}
	list := []workflow_support.Worker{&worker}
	scheduler := workflow_support.WorkScheduler{Finder: &finder, Workers: list, Delay: 50}
	scheduler.Start()
	time.Sleep(time.Duration(50) * time.Millisecond)
	scheduler.Stop()

	assert.Equal(t, finder.Completed, 0)
}

type ErroneousWorker struct {
}

func (n *ErroneousWorker) Run(interface{}) error {
	return errors.New("oops")
}

func TestWorkflow_bad_find(t *testing.T) {
	gateway := setUp()
	_, _ = gateway.Create("aName", "google cloud", []byte("aKey"))

	worker := ErroneousWorker{}
	finder := orchestrator.DiscoveryWorkFinder{Gateway: gateway}
	list := []workflow_support.Worker{&worker}
	scheduler := workflow_support.WorkScheduler{Finder: &finder, Workers: list, Delay: 50}
	scheduler.Start()
	for finder.NotCompleted < 1 {
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
	scheduler.Stop()

	assert.True(t, finder.NotCompleted > 0)
}
