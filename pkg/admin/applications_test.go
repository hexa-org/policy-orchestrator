package admin_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"hexa/pkg/admin"
	"hexa/pkg/admin/test"
	"hexa/pkg/web_support"
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
