package orchestrator_test

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"net"
	"net/http"
	"testing"
)

type ApplicationsHandlerSuite struct {
	suite.Suite
	fields *orchestrator_test.SuiteFields
}

func TestApplicationsHandler(t *testing.T) {
	suite.Run(t, &ApplicationsHandlerSuite{fields: &orchestrator_test.SuiteFields{}})
}

func (suite *ApplicationsHandlerSuite) SetupTest() {
	listener, _ := net.Listen("tcp", "localhost:0")
	suite.fields.Setup(listener.Addr().String())
	go websupport.Start(suite.fields.Server, listener)
	websupport.WaitForHealthy(suite.fields.Server)
}

func (suite *ApplicationsHandlerSuite) TearDownTest() {
	_ = suite.fields.DB.Close()
	websupport.Stop(suite.fields.Server)
}

func (suite *ApplicationsHandlerSuite) TestList() {
	var integrationTestId string
	_ = suite.fields.DB.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "aProvider", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = suite.fields.DB.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	resp, err := hawksupport.HawkGet(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://%s/applications", suite.fields.Server.Addr))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var apps orchestrator.Applications
	_ = json.NewDecoder(resp.Body).Decode(&apps)
	assert.Equal(suite.T(), 1, len(apps.Applications))
	assert.Equal(suite.T(), "anObjectId", apps.Applications[0].ObjectId)
	assert.Equal(suite.T(), "aName", apps.Applications[0].Name)
	assert.Equal(suite.T(), "aDescription", apps.Applications[0].Description)
}

func (suite *ApplicationsHandlerSuite) TestShow() {
	var integrationTestId string
	_ = suite.fields.DB.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "aProvider", []byte("aKey")).Scan(&integrationTestId)

	var applicationTestId string
	_ = suite.fields.DB.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&applicationTestId)

	resp, err := hawksupport.HawkGet(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://%s/applications/%s", suite.fields.Server.Addr, applicationTestId))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var app orchestrator.Application
	_ = json.NewDecoder(resp.Body).Decode(&app)
	assert.Equal(suite.T(), "anObjectId", app.ObjectId)
	assert.Equal(suite.T(), "aName", app.Name)
	assert.Equal(suite.T(), "aDescription", app.Description)
}

func (suite *ApplicationsHandlerSuite) TestShow_identifier() {
	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://%s/applications/oops", suite.fields.Server.Addr))
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)
}
