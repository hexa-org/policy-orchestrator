package orchestrator_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

type integrationsTestData struct {
	gateway orchestrator.IntegrationsDataGateway
}

func (data *integrationsTestData) SetUp() {
	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	data.gateway = orchestrator.IntegrationsDataGateway{DB: db}
	_, _ = db.Exec("delete from integrations;")
}

func (data *integrationsTestData) TearDown() {
	_ = data.gateway.DB.Close()
}

func TestCreateIntegration(t *testing.T) {
	testsupport.WithSetUp(&integrationsTestData{}, func(data *integrationsTestData) {
		id, err := data.gateway.Create("aName", "google_cloud", []byte("aKey"))
		assert.NotEmpty(t, id)
		assert.NoError(t, err)
	})
}

func TestFindIntegrations(t *testing.T) {
	testsupport.WithSetUp(&integrationsTestData{}, func(data *integrationsTestData) {
		_, _ = data.gateway.Create("aName", "google_cloud", []byte("aKey"))
		all, _ := data.gateway.Find()
		assert.Equal(t, 1, len(all))
	})
}

func TestFindIntegrations_withDatabaseError(t *testing.T) {
	testsupport.WithSetUp(&integrationsTestData{}, func(data *integrationsTestData) {
		open, _ := databasesupport.Open("")

		gateway := orchestrator.IntegrationsDataGateway{DB: open}
		_, _ = data.gateway.Create("aName", "google_cloud", []byte("aKey"))
		_, err := gateway.Find()
		assert.Error(t, err)
	})
}

func TestDeleteIntegration(t *testing.T) {
	testsupport.WithSetUp(&integrationsTestData{}, func(data *integrationsTestData) {
		id, _ := data.gateway.Create("aName", "google_cloud", []byte("aKey"))
		_ = data.gateway.Delete(id)
		find, _ := data.gateway.Find()
		assert.Equal(t, 0, len(find))
	})
}

func TestFindIntegrationById(t *testing.T) {
	testsupport.WithSetUp(&integrationsTestData{}, func(data *integrationsTestData) {
		id, _ := data.gateway.Create("aName", "google_cloud", []byte("aKey"))
		record, _ := data.gateway.FindById(id)
		assert.Equal(t, "aName", record.Name)
		assert.Equal(t, "google_cloud", record.Provider)
		assert.Equal(t, []byte("aKey"), record.Key)
	})
}
