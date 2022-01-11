package admin_test

import (
	"github.com/stretchr/testify/assert"
	"hexa/pkg/admin"
	admin_test "hexa/pkg/admin/test"
	"hexa/pkg/web_support"
	"net/http"
	"testing"
)

func TestAdminHandlers(t *testing.T) {
	handlers := admin.LoadHandlers("localhost:8885", new(admin_test.MockClient))
	server := web_support.Create("localhost:8883", handlers, web_support.Options{})
	go web_support.Start(server)
	web_support.WaitForHealthy(server)

	resp, _ := http.Get("http://localhost:8883/health")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	noFollowClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	redirect, _ := noFollowClient.Get("http://localhost:8883")
	assert.Equal(t, http.StatusPermanentRedirect, redirect.StatusCode)
	assert.Equal(t, string(redirect.Header["Location"][0]), "/integrations")

	web_support.Stop(server)
}
