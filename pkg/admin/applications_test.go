package admin_test

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"net"
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

	listener, _ := net.Listen("tcp", "localhost:0")
	suite.client = new(admin_test.MockClient)
	suite.server = websupport.Create(
		listener.Addr().String(),
		admin.LoadHandlers("http://noop", suite.client),
		websupport.Options{ResourceDirectory: resourcesDirectory})
	go websupport.Start(suite.server, listener)
	websupport.WaitForHealthy(suite.server)
}

func (suite *ApplicationsSuite) TearDownTest() {
	websupport.Stop(suite.server)
}

///

func (suite *ApplicationsSuite) TestApplications() {
	resp, err := http.Get(fmt.Sprintf("http://%s/applications", suite.server.Addr))
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
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}

func (suite *ApplicationsSuite) TestApplication() {
	identifier := "anId"
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/%s", suite.server.Addr, identifier))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Applications")
	assert.Contains(suite.T(), string(body), "anObjectId")
	assert.Contains(suite.T(), string(body), "aName")
	assert.Contains(suite.T(), string(body), "aDescription")
}

func (suite *ApplicationsSuite) TestApplication_with_error() {
	suite.client.Err = errors.New("oops")
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/000", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}
