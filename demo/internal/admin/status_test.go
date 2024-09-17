package admin_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/demo/internal/admin"
	"github.com/hexa-org/policy-orchestrator/demo/internal/admin/test"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/pkg/healthsupport"
	"github.com/hexa-org/policy-mapper/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

type StatusData struct {
	server *http.Server
	client adminMock.MockClient
}

func (data *StatusData) SetUp() {
	data.client = adminMock.MockClient{}
	handler := admin.NewStatusHandler("http://noop", &data.client)
	listener, _ := net.Listen("tcp", "localhost:0")
	data.server = websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/status", handler.StatusHandler).Methods("GET")
	}, websupport.Options{})

	go websupport.Start(data.server, listener)
	healthsupport.WaitForHealthy(data.server)
}

func (data *StatusData) TearDown() {
	websupport.Stop(data.server)
	data.server = nil
}

func TestStatusHandler(t *testing.T) {
	testsupport.WithSetUp(&StatusData{}, func(data *StatusData) {
		data.client.Status = "[{\"name\":\"noop\",\"pass\":\"true\"}]"

		resp, _ := http.Get(fmt.Sprintf("http://%s/status", data.server.Addr))
		body, _ := io.ReadAll(resp.Body)

		assert.Contains(t, string(body), "Hexa Policy Orchestrator Status")
		assert.Contains(t, string(body), "http://noop")
		assert.Contains(t, string(body), "<a class=\"status green\">")
	})
}

func TestStatusHandler_withBadResponse(t *testing.T) {
	testsupport.WithSetUp(&StatusData{}, func(data *StatusData) {
		data.client.Status = "x"

		resp, _ := http.Get(fmt.Sprintf("http://%s/status", data.server.Addr))
		body, _ := io.ReadAll(resp.Body)

		assert.Contains(t, string(body), "Hexa Policy Orchestrator Status")
		assert.Contains(t, string(body), "http://noop")
		assert.Contains(t, string(body), "<a class=\"status orange\">")
	})
}
