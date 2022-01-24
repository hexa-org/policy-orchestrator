package orchestrator_test

import (
	"database/sql"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type GatewaySuite struct {
	suite.Suite
	db      *sql.DB
	gateway orchestrator.IntegrationsDataGateway
}

func TestIntegrationsDataGateway(t *testing.T) {
	suite.Run(t, new(GatewaySuite))
}

func (suite *GatewaySuite) SetupTest() {
	suite.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	suite.gateway = orchestrator.IntegrationsDataGateway{DB: suite.db}
	_, _ = suite.db.Exec("delete from integrations;")
}

func (suite *GatewaySuite) TearDownTest() {
	_ = suite.db.Close()
}

func (suite *GatewaySuite) TestCreate() {
	id, err := suite.gateway.Create("aName", "google cloud", []byte("aKey"))
	assert.NotEmpty(suite.T(), id)
	assert.NoError(suite.T(), err)
}

func (suite *GatewaySuite) TestFind() {
	_, _ = suite.gateway.Create("aName", "google cloud", []byte("aKey"))
	all, _ := suite.gateway.Find()
	assert.Equal(suite.T(), 1, len(all))
}

func (suite *GatewaySuite) TestFind_withBadDatabaseUrl() {
	open, _ := databasesupport.Open("")
	gateway := orchestrator.IntegrationsDataGateway{DB: open}
	_, _ = suite.gateway.Create("aName", "google cloud", []byte("aKey"))
	_, err := gateway.Find()
	assert.Error(suite.T(), err)
}

func (suite *GatewaySuite) TestDelete() {
	id, _ := suite.gateway.Create("aName", "google cloud", []byte("aKey"))
	_ = suite.gateway.Delete(id)
	find, _ := suite.gateway.Find()
	assert.Equal(suite.T(), 0, len(find))
}

func (suite *GatewaySuite) TestFindById() {
	id, _ := suite.gateway.Create("aName", "google cloud", []byte("aKey"))
	record, _ := suite.gateway.FindById(id)
	assert.Equal(suite.T(), "aName", record.Name)
	assert.Equal(suite.T(), "google cloud", record.Provider)
	assert.Equal(suite.T(), []byte("aKey"), record.Key)
}
