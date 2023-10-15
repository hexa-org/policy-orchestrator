package orchestrator_test

import (
	"errors"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"
	"testing"
	"time"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/workflowsupport"
	"github.com/stretchr/testify/assert"
)

func setUp() (orchestrator.IntegrationsDataGateway, orchestrator.ApplicationsDataGateway) {
	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = db.Exec("delete from integrations;")
	integrationsGateway := orchestrator.IntegrationsDataGateway{DB: db}
	appGateway := orchestrator.ApplicationsDataGateway{DB: db}
	return integrationsGateway, appGateway
}

func TestWorkflow(t *testing.T) {
	integrationsGateway, appGateway := setUp()
	_, _ = integrationsGateway.Create("aName", "noop", []byte("aKey"))

	noopProvider := orchestrator_test.NoopProvider{}
	providers := make(map[string]orchestrator.Provider)
	providers["noop"] = &noopProvider
	worker := orchestrator.DiscoveryWorker{Providers: providers, Gateway: appGateway}
	finder := orchestrator.NewDiscoveryWorkFinder(integrationsGateway)
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)

	scheduler.Start()
	assert.True(t, <-finder.Results)
	scheduler.Stop()

	find, _ := appGateway.Find()
	assert.Equal(t, 3, len(find))
	assert.True(t, noopProvider.Discovered > 2)
}

func TestRemoveDeletedApplications(t *testing.T) {
	integrationsGateway, appDataGateway := setUp()

	id, _ := integrationsGateway.Create("aName", "noop", []byte("aKey"))
	_, _ = appDataGateway.CreateIfAbsent(id, "object1", "app1", "", "service1")
	app2ID, _ := appDataGateway.CreateIfAbsent(id, "object2", "app2", "", "service2")

	provider := fakeProvider{
		discoveredApplications: []orchestrator.ApplicationInfo{
			{
				ObjectID: "object2",
				Name:     "app2",
				Service:  "service2",
			},
		},
	}
	discoveryWorker := &orchestrator.DiscoveryWorker{
		Providers: map[string]orchestrator.Provider{
			"fake": provider,
		},
		Gateway: appDataGateway,
	}
	work := []orchestrator.IntegrationRecord{{Provider: "fake"}}

	discoveryWorker.Run(work)

	found, err := appDataGateway.Find()
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, app2ID, found[0].ID)
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
	_, _ = gateway.Create("aName", "noop", []byte("aKey"))

	worker := ErroneousWorker{}
	finder := orchestrator.NewDiscoveryWorkFinder(gateway)
	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)

	scheduler.Start()
	assert.False(t, <-finder.Results)
	scheduler.Stop()
}

type fakeProvider struct {
	discoveredApplications []orchestrator.ApplicationInfo
}

func (f fakeProvider) Name() string {
	return "fake"
}

func (f fakeProvider) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	return f.discoveredApplications, nil
}

func (f fakeProvider) GetPolicyInfo(info orchestrator.IntegrationInfo, info2 orchestrator.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	panic("implement me")
}

func (f fakeProvider) SetPolicyInfo(info orchestrator.IntegrationInfo, info2 orchestrator.ApplicationInfo, infos []hexapolicy.PolicyInfo) (status int, foundErr error) {
	panic("implement me")
}
