package orchestrator_test

import (
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	assert "github.com/stretchr/testify/assert"
)

type applicationsTestData struct {
	integrationTestId string
	gateway           orchestrator.ApplicationsDataGateway
}

func (data *applicationsTestData) SetUp() {
	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	data.gateway = orchestrator.ApplicationsDataGateway{DB: db}
	data.integrationTestId = "50e00619-9f15-4e85-a7e9-f26d87ea12e7"
	_, _ = db.Exec(`
delete from applications;
delete from integrations;
insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'aName', 'noop', 'aKey');
`)
}

func (data *applicationsTestData) TearDown() {
	_ = data.gateway.DB.Close()
}

func TestCreateApp(t *testing.T) {
	testsupport.WithSetUp(&applicationsTestData{}, func(data *applicationsTestData) {
		id, err := data.gateway.CreateIfAbsent(data.integrationTestId, "anObjectId", "aName", "aDescription", "aService")
		assert.NotEmpty(t, id)
		assert.NoError(t, err)
	})
}

func TestFindApps(t *testing.T) {
	testsupport.WithSetUp(&applicationsTestData{}, func(data *applicationsTestData) {
		_, err := data.gateway.CreateIfAbsent(data.integrationTestId, "anObjectId", "aName", "aDescription", "aService")
		assert.NoError(t, err)
		records, _ := data.gateway.Find()
		record := records[0]
		assert.Equal(t, data.integrationTestId, record.IntegrationId)
		assert.Equal(t, "anObjectId", record.ObjectId)
		assert.Equal(t, "aName", record.Name)
		assert.Equal(t, "aDescription", record.Description)
		assert.Equal(t, "aService", record.Service)
	})
}

func TestFindByIntegrationId(t *testing.T) {
	testsupport.WithSetUp(&applicationsTestData{}, func(data *applicationsTestData) {
		_, err := data.gateway.CreateIfAbsent(data.integrationTestId, "anObjectId", "aName", "aDescription", "aService")
		assert.NoError(t, err)
		record, _ := data.gateway.FindByIntegrationId(data.integrationTestId)
		assert.Equal(t, data.integrationTestId, record.IntegrationId)
		assert.Equal(t, "anObjectId", record.ObjectId)
		assert.Equal(t, "aName", record.Name)
		assert.Equal(t, "aDescription", record.Description)
		assert.Equal(t, "aService", record.Service)
	})
}

func TestFindApps_withDatabaseError(t *testing.T) {
	open, _ := databasesupport.Open("")
	gateway := orchestrator.ApplicationsDataGateway{DB: open}
	_, _ = gateway.CreateIfAbsent("anId", "anObjectId", "aName", "aDescription", "aService")
	_, err := gateway.Find()
	assert.Error(t, err)
}

func TestFindApps_ignoresDuplicates(t *testing.T) {
	testsupport.WithSetUp(&applicationsTestData{}, func(data *applicationsTestData) {
		_, _ = data.gateway.CreateIfAbsent(data.integrationTestId, "anObjectId", "aName", "aDescription", "aService")
		_, _ = data.gateway.CreateIfAbsent(data.integrationTestId, "anObjectId", "aName", "aDescription", "aService")
		find, _ := data.gateway.Find()
		assert.Equal(t, 1, len(find))
	})
}

func TestFindAppById(t *testing.T) {
	testsupport.WithSetUp(&applicationsTestData{}, func(data *applicationsTestData) {
		id, _ := data.gateway.CreateIfAbsent(data.integrationTestId, "anObjectId", "aName", "aDescription", "aService")
		record, err := data.gateway.FindById(id)
		assert.NoError(t, err)
		assert.Equal(t, data.integrationTestId, record.IntegrationId)
		assert.Equal(t, "anObjectId", record.ObjectId)
		assert.Equal(t, "aName", record.Name)
		assert.Equal(t, "aDescription", record.Description)
		assert.Equal(t, "aService", record.Service)
	})
}
