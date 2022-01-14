package admin_test

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
)

type ApplicationsSuite struct {
	suite.Suite
	server *http.Server
	client *admin_test.MockClient
}

func TestApplications(t *testing.T) {
	suite.Run(t, new(ApplicationsSuite))
}

func (suite *ApplicationsSuite) SetupTest() {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")

	suite.client = new(admin_test.MockClient)
	suite.server = web_support.Create(
		"localhost:8883",
		admin.LoadHandlers("http://noop", suite.client),
		web_support.Options{ResourceDirectory: resourcesDirectory})
	go web_support.Start(suite.server)
	web_support.WaitForHealthy(suite.server)
}

func (suite *ApplicationsSuite) TearDownTest() {
	web_support.Stop(suite.server)
}

///

func (suite *ApplicationsSuite) TestApplications() {
	resp, err := http.Get("http://localhost:8883/applications")
	if err != nil {
		log.Fatalln(err)
	}
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Applications")
	assert.Contains(suite.T(), string(body), "anObjectId")
	assert.Contains(suite.T(), string(body), "aName")
	assert.Contains(suite.T(), string(body), "aDescription")
}

func (suite *ApplicationsSuite) TestApplications_with_error() {
	suite.client.Err = errors.New("oops")
	resp, _ := http.Get("http://localhost:8883/applications")
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}

func (suite *ApplicationsSuite) TestApplication() {
	identifier := "anId"
	resp, _ := http.Get(fmt.Sprintf("http://localhost:8883/applications/%s", identifier))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Applications")
	assert.Contains(suite.T(), string(body), "anObjectId")
	assert.Contains(suite.T(), string(body), "aName")
	assert.Contains(suite.T(), string(body), "aDescription")
}

func (suite *ApplicationsSuite) TestApplication_with_error() {
	suite.client.Err = errors.New("oops")
	resp, _ := http.Get("http://localhost:8883/applications/000")
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}