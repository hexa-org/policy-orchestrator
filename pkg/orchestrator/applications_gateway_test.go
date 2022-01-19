package orchestrator_test

import (
	"database/sql"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ApplicationGatewaySuite struct {
	suite.Suite
	db                *sql.DB
	gateway           orchestrator.ApplicationsDataGateway
	integrationTestId string
}

func TestApplicationsDataGateway(t *testing.T) {
	suite.Run(t, new(ApplicationGatewaySuite))
}

func (suite *ApplicationGatewaySuite) SetupTest() {
	suite.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = suite.db.Exec("delete from applications;")
	_, _ = suite.db.Exec("delete from integrations;")
	_ = suite.db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "aProvider", []byte("aKey")).Scan(&suite.integrationTestId)

	suite.gateway = orchestrator.ApplicationsDataGateway{DB: suite.db}
}

func (suite *ApplicationGatewaySuite) TearDownTest() {
	_ = suite.db.Close()
}

///

func (suite *ApplicationGatewaySuite) TestCreate() {
	id, err := suite.gateway.Create(suite.integrationTestId, "anObjectId", "aName", "aDescription")
	assert.NotEmpty(suite.T(), id)
	assert.NoError(suite.T(), err)
}

func (suite *ApplicationGatewaySuite) TestFind() {
	_, _ = suite.gateway.Create(suite.integrationTestId, "anObjectId", "aName", "aDescription")
	all, _ := suite.gateway.Find()
	assert.Equal(suite.T(), suite.integrationTestId, all[0].IntegrationId)
	assert.Equal(suite.T(), "anObjectId", all[0].ObjectId)
	assert.Equal(suite.T(), "aName", all[0].Name)
	assert.Equal(suite.T(), "aDescription", all[0].Description)
}

func (suite *ApplicationGatewaySuite) TestFind_withBadDatabaseUrl() {
	open, _ := databasesupport.Open("")
	gateway := orchestrator.ApplicationsDataGateway{DB: open}
	_, _ = gateway.Create(suite.integrationTestId, "anObjectId", "aName", "aDescription")
	_, err := gateway.Find()
	assert.Error(suite.T(), err)
}

func (suite *ApplicationGatewaySuite) TestFind_ignoresDuplicates() {
	_, _ = suite.gateway.Create(suite.integrationTestId, "anObjectId", "aName", "aDescription")
	_, _ = suite.gateway.Create(suite.integrationTestId, "anObjectId", "aName", "aDescription")
	find, _ := suite.gateway.Find()
	assert.Equal(suite.T(), 1, len(find))
}

func (suite *ApplicationGatewaySuite) TestFindById() {
	id, _ := suite.gateway.Create(suite.integrationTestId, "anObjectId", "aName", "aDescription")
	found, _ := suite.gateway.FindById(id)
	assert.Equal(suite.T(), suite.integrationTestId, found.IntegrationId)
	assert.Equal(suite.T(), "anObjectId", found.ObjectId)
	assert.Equal(suite.T(), "aName", found.Name)
	assert.Equal(suite.T(), "aDescription", found.Description)
}
