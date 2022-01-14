package admin_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
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

type StatusSuite struct {
	suite.Suite
	server *http.Server
}

func TestStatus(t *testing.T) {
	suite.Run(t, new(StatusSuite))
}

func (suite *StatusSuite) SetupTest() {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")

	handler := admin.NewStatusHandler("http://noop", new(admin_test.MockClient))
	listener, _ := net.Listen("tcp", "localhost:0")
	suite.server = web_support.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/status", handler.StatusHandler).Methods("GET")
	}, web_support.Options{ResourceDirectory: resourcesDirectory})

	go web_support.Start(suite.server, listener)
	web_support.WaitForHealthy(suite.server)
}

func (suite *StatusSuite) TearDownTest() {
	web_support.Stop(suite.server)
}

///

func (suite *StatusSuite) TestStatus() {
	resp, err := http.Get(fmt.Sprintf("http://%s/status", suite.server.Addr))
	if err != nil {
		log.Fatalln(err)
	}
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Hexa Policy Orchestrator Status")
	assert.Contains(suite.T(), string(body), "http://noop")
}
