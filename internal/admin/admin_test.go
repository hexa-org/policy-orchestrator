package admin_test

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/admin"
	"github.com/hexa-org/policy-orchestrator/internal/admin/test"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestAdminHandlers(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	handlers := admin.LoadHandlers("localhost:8885", new(admin_test.MockClient))
	server := websupport.Create(listener.Addr().String(), handlers, websupport.Options{})
	go websupport.Start(server, listener)
	healthsupport.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	noFollowClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	redirect, _ := noFollowClient.Get(fmt.Sprintf("http://%s", server.Addr))
	assert.Equal(t, http.StatusPermanentRedirect, redirect.StatusCode)
	assert.Equal(t, redirect.Header["Location"][0], "/integrations")

	websupport.Stop(server)
}
