package orchestrator_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
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

	discovery := orchestrator_test.NoopDiscovery{}
	worker := orchestrator.DiscoveryWorker{Providers: []provider.Provider{&discovery}, Gateway: appGateway}
	finder := orchestrator.DiscoveryWorkFinder{Gateway: gateway}
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()
	for atomic.LoadInt32(&finder.Completed) < 1 {
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
	scheduler.Stop()

	find, _ := appGateway.Find()
	assert.Equal(t, 3, len(find))
	assert.True(t, atomic.LoadInt32(&finder.Completed) > 0)
	assert.True(t, discovery.Discovered > 2)
}

func TestWorkflow_empty(t *testing.T) {
	gateway, _ := setUp()

	worker := orchestrator.DiscoveryWorker{}
	finder := orchestrator.DiscoveryWorkFinder{Gateway: gateway}
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()
	time.Sleep(time.Duration(50) * time.Millisecond)
	scheduler.Stop()

	assert.Equal(t, atomic.LoadInt32(&finder.Completed), int32(0))
}

type ErroneousWorker struct {
}

func (n *ErroneousWorker) Run(interface{}) error {
	return errors.New("oops")
}

func TestWorkflow_bad_find(t *testing.T) {
	gateway, _ := setUp()
	_, _ = gateway.Create("aName", "google cloud", []byte("aKey"))

	worker := ErroneousWorker{}
	finder := orchestrator.DiscoveryWorkFinder{Gateway: gateway}
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()
	for atomic.LoadInt32(&finder.NotCompleted) < 1 {
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
	scheduler.Stop()

	assert.True(t, atomic.LoadInt32(&finder.NotCompleted) > 0)
}
