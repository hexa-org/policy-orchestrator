//go:build integration

package demosmoke_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

type Integration struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Key      []byte `json:"key"`
}

type Applications struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	ID string `json:"id"`
}

type Policy struct {
	Version string  `json:"version"`
	Action  string  `json:"action"`
	Subject Subject `json:"subject"`
	Object  Object  `json:"object"`
}

type Subject struct {
	AuthenticatedUsers []string `json:"authenticated_users"`
}

type Object struct {
	Resources []string `json:"resources"`
}

func TestDemoFlow(t *testing.T) {

	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = db.Exec("delete from integrations; delete from applications;")

	demo := makeGoCommand("/cmd/demo/demo.go", []string{"PORT=8886", "OPA_SERVER_URL: http://localhost:8887/v1/data/authz/allow"}, []string{})
	demoConfig := makeGoCommand("/cmd/democonfig/democonfig.go", []string{"PORT=8889"}, []string{})
	demoProxy := makeGoCommand("/cmd/demoproxy/demoproxy.go", []string{"PORT=8890", "REMOTER_URL=http://localhost:8886"}, []string{""})
	orchestrator := makeGoCommand("/cmd/orchestrator/orchestrator.go", []string{
		"PORT=8885",
		"ORCHESTRATOR_HOSTPORT=localhost:8885",
		"ORCHESTRATOR_KEY=0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb",
		"POSTGRESQL_URL=postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable",
	}, []string{""})

	_, file, _, _ := runtime.Caller(0)
	config := filepath.Join(file, "../../../deployments/opa-server/config/config.yaml")
	openPolicyAgent := exec.Command("opa", "run", "--server", "--addr", "localhost:8887", "-c", config)
	openPolicyAgent.Env = os.Environ()
	openPolicyAgent.Env = append(openPolicyAgent.Env, "HEXA_DEMO_URL=http://localhost:8889")

	start(demo, 8886)
	start(demoConfig, 8889)
	start(demoProxy, 8890)
	start(openPolicyAgent, 8887)
	start(orchestrator, 8885)

	defer func() {
		stopAll(orchestrator, openPolicyAgent, demoProxy, demoConfig, demo)
	}()

	resp, _ := http.Get("http://localhost:8890/")
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Great news, you're able to access this page.")

	resp, _ = http.Get("http://localhost:8890/sales")
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Great news, you're able to access this page.")

	resp, _ = http.Get("http://localhost:8890/accounting")
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Sorry, you're not able to access this page.")

	resp, _ = http.Get("http://localhost:8890/humanresources")
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Sorry, you're not able to access this page.")

	integrationInfo, _ := json.Marshal(Integration{Name: "bundle:open-policy-agent", Provider: "open_policy_agent",
		Key: []byte(`{ "bundle_url":"http://localhost:8889/bundles/bundle.tar.gz" }`)})

	resp, _ = orchestratorPost("http://localhost:8885/integrations", bytes.NewReader(integrationInfo))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, _ = orchestratorGet("http://localhost:8885/applications")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var apps Applications
	_ = json.NewDecoder(resp.Body).Decode(&apps)

	var policies bytes.Buffer
	policy := Policy{Version: "v0.4", Action: "GET",
		Subject: Subject{AuthenticatedUsers: []string{"accounting@hexaindustries.io", "sales@hexaindustries.io"}},
		Object:  Object{Resources: []string{"/accounting"}},
	}
	_ = json.NewEncoder(&policies).Encode([]Policy{policy})

	resp, _ = orchestratorPost(fmt.Sprintf("http://localhost:8885/applications/%s/policies",
		apps.Applications[0].ID), bytes.NewReader(policies.Bytes()))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	time.Sleep(time.Duration(30) * time.Second) // waiting for opa to refresh the bundle

	resp, _ = http.Get("http://localhost:8890/accounting")
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Great news, you're able to access this page.")
}

/// supporting functions

func makeGoCommand(cmdString string, envs []string, args []string) *exec.Cmd {
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(file, "../../../")
	commandPath := filepath.Join(path, cmdString)

	args = append([]string{commandPath}, args...)
	args = append([]string{"run"}, args...)

	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, envs...)

	// assigning parent and child processes to a process group to ensure all process receive stop signal
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}

func start(cmd *exec.Cmd, port int) {
	log.Printf("Starting cmd %v\n", cmd)
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Printf("Unable to start cmd %v\n.", err)
		}
	}()
	waitForHealthy(fmt.Sprintf("localhost:%v", port))
}

func waitForHealthy(address string) {
	var isLive bool
	for !isLive {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", address))
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Println("Server is healthy.", address)
			isLive = true
		}
	}
}

func orchestratorGet(url string) (*http.Response, error) {
	return hawksupport.HawkGet(&http.Client{},
		"anId", "0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb", url)
}

func orchestratorPost(url string, reader *bytes.Reader) (*http.Response, error) {
	return hawksupport.HawkPost(&http.Client{},
		"anId", "0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb", url, reader)
}

func stopAll(cmds ...*exec.Cmd) {
	for _, cmd := range cmds {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
