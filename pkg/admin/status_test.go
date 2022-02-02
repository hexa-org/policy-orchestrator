package admin_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
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
	suite.server = websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/status", handler.StatusHandler).Methods("GET")
	}, websupport.Options{ResourceDirectory: resourcesDirectory})

	go websupport.Start(suite.server, listener)
	healthsupport.WaitForHealthy(suite.server)
}

func (suite *StatusSuite) TearDownTest() {
	websupport.Stop(suite.server)
}

func (suite *StatusSuite) TestStatus() {
	resp, _ := http.Get(fmt.Sprintf("http://%s/status", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Hexa Policy Orchestrator Status")
	assert.Contains(suite.T(), string(body), "http://noop")
}
