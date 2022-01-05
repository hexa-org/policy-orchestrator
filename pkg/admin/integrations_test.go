package admin_test

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"hexa/pkg/admin"
	"hexa/pkg/web_support"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
)

type IntegrationsSuite struct {
	suite.Suite
	server *http.Server
}

func TestIntegrations(t *testing.T) {
	suite.Run(t, new(IntegrationsSuite))
}

func (suite *IntegrationsSuite) SetupTest() {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")

	handler := admin.NewIntegrationsHandler()
	suite.server = web_support.Create("localhost:8883", func(router *mux.Router) {
		router.HandleFunc("/discovery", handler.IntegrationsHandler).Methods("GET")
	}, web_support.Options{ResourceDirectory: resourcesDirectory})

	go web_support.Start(suite.server)
	web_support.WaitForHealthy(suite.server)
}

func (suite *IntegrationsSuite) TearDownTest() {
	web_support.Stop(suite.server)
}

///

func (suite *IntegrationsSuite) TestIntegrations() {
	resp, err := http.Get("http://localhost:8883/discovery")
	if err != nil {
		log.Fatalln(err)
	}
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Discovery")
}
