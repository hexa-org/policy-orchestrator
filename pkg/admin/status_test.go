package admin_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
)

type StatusData struct {
	server *http.Server
}

func (data *StatusData) SetUp() {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")

	handler := admin.NewStatusHandler("http://noop", new(admin_test.MockClient))
	listener, _ := net.Listen("tcp", "localhost:0")
	data.server = websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/status", handler.StatusHandler).Methods("GET")
	}, websupport.Options{ResourceDirectory: resourcesDirectory})

	go websupport.Start(data.server, listener)
	healthsupport.WaitForHealthy(data.server)
}

func (data *StatusData) TearDown() {
	websupport.Stop(data.server)
}

func TestNewStatusHandler(t *testing.T) {
	testsupport.WithSetUp(&StatusData{}, func(data *StatusData) {
		resp, _ := http.Get(fmt.Sprintf("http://%s/status", data.server.Addr))
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Hexa Policy Orchestrator Status")
		assert.Contains(t, string(body), "http://noop")
	})
}
