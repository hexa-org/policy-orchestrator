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

func TestDemoFlow(t *testing.T) {

	db, _ := databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	deleteAll := "delete from integrations; delete from applications;"

	demo := makeCmd("/cmd/demo/demo.go", []string{"PORT=8886", "OPA_SERVER_URL: http://localhost:8887/v1/data/authz/allow"})
	demoConfig := makeCmd("/cmd/democonfig/democonfig.go", []string{"PORT=8889"})
	orchestrator := makeCmd("/cmd/orchestrator/orchestrator.go", []string{
		"PORT=8885",
		"ORCHESTRATOR_HOSTPORT=localhost:8885",
		"ORCHESTRATOR_KEY=0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb",
		"POSTGRESQL_URL=postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable",
	})

	_, file, _, _ := runtime.Caller(0)
	testBundles := filepath.Join(file, "../../../cmd/democonfig/resources/bundles/.bundle-*")
	files, _ := filepath.Glob(testBundles)
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			panic(err)
		}
	}
	config := filepath.Join(file, "../../../cmd/demosmoke/config.yaml")
	openPolicyAgent := exec.Command("opa", "run", "--server", "--addr", "localhost:8887", "-c", config)
	openPolicyAgent.Env = os.Environ()
	openPolicyAgent.Env = append(openPolicyAgent.Env, "HEXA_DEMO_URL=http://localhost:8889")
	openPolicyAgent.Stdout = os.Stdout
	openPolicyAgent.Stderr = os.Stderr
	openPolicyAgent.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	startCmd(demo, 8886)
	startCmd(demoConfig, 8889)
	startCmd(openPolicyAgent, 8887)
	startCmd(orchestrator, 8885)

	defer func() {
		stopCmds(orchestrator, openPolicyAgent, demoConfig, demo)
	}()

	assertContains(t, "http://localhost:8886/", "Great news, you're able to access this page.")

	assertContains(t, "http://localhost:8886/sales", "Great news, you're able to access this page.")

	assertContains(t, "http://localhost:8886/accounting", "Sorry, you're not able to access this page.")

	assertContains(t, "http://localhost:8886/humanresources", "Sorry, you're not able to access this page.")

	_, _ = db.Exec(deleteAll)
	createAnIntegration()

	status, updateErr := updateAPolicy()
	assert.Equal(t, http.StatusCreated, status.StatusCode)
	assert.NoError(t, updateErr)

	time.Sleep(time.Duration(3) * time.Second) // waiting for opa to refresh the bundle

	assertContains(t, "http://localhost:8886/accounting", "Great news, you're able to access this page.")

	_, _ = db.Exec(deleteAll)
	createAnErroneousIntegration()

	resp, secondUpdateErr := updateAPolicy()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.NoError(t, secondUpdateErr)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "unable to update policy.\n", string(body))

	_, _ = http.Get("http://localhost:8889/reset")
}

func assertContains(t *testing.T, url string, contains string) {
	resp, _ := http.Get(url)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), contains, url)
}

func createAnIntegration() {
	integrationInfo, _ := json.Marshal(Integration{Name: "bundle:open-policy-agent", Provider: "open_policy_agent",
		Key: []byte(`{ "bundle_url":"http://localhost:8889/bundles/bundle.tar.gz" }`)})

	_, _ = hawksupport.HawkPost(&http.Client{},
		"anId", "0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb",
		"http://localhost:8885/integrations", bytes.NewReader(integrationInfo))
}

func createAnErroneousIntegration() {
	integrationInfo, _ := json.Marshal(Integration{Name: "bundle:open-policy-agent", Provider: "open_policy_agent",
		Key: []byte(`{ "bundle_url":"http://localhost:8800" }`)})

	_, _ = hawksupport.HawkPost(&http.Client{},
		"anId", "0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb",
		"http://localhost:8885/integrations", bytes.NewReader(integrationInfo))
}

func updateAPolicy() (*http.Response, error) {
	var apps Applications
	resp, _ := hawksupport.HawkGet(&http.Client{},
		"anId", "0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb",
		"http://localhost:8885/applications")
	_ = json.NewDecoder(resp.Body).Decode(&apps)

	var policies bytes.Buffer
	policy := Policy{Meta: Meta{"0.5"}, Actions: []Action{{"http:GET"}},
		Subject: Subject{Members: []string{"accounting@hexaindustries.io", "sales@hexaindustries.io"}},
		Object:  Object{Resources: []string{"/accounting"}},
	}
	_ = json.NewEncoder(&policies).Encode(Policies{[]Policy{policy}})

	url := fmt.Sprintf("http://localhost:8885/applications/%s/policies", apps.Applications[0].ID)
	resp, err := hawksupport.HawkPost(&http.Client{},
		"anId", "0861f51ab66590798406be5b184c71b637bfc907c83f27d461e4956bffebf6cb",
		url, bytes.NewReader(policies.Bytes()))
	return resp, err
}

/// supporting structs

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

type Policies struct {
	Policies []Policy `json:"policies"`
}

type Policy struct {
	Meta    Meta     `json:"meta"`
	Actions []Action `json:"actions"`
	Subject Subject  `json:"subject"`
	Object  Object   `json:"object"`
}

type Meta struct {
	Version string `json:"version"`
}

type Action struct {
	Action string `json:"action"`
}

type Subject struct {
	Members []string `json:"members"`
}

type Object struct {
	Resources []string `json:"resources"`
}

/// supporting functions

func makeCmd(cmdString string, envs []string) *exec.Cmd {
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(file, "../../../")
	commandPath := filepath.Join(path, cmdString)

	var args []string
	args = append([]string{commandPath}, args...)
	args = append([]string{"run"}, args...)

	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, envs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// assigning parent and child processes to a process group to ensure all process receive stop signal
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}

func startCmd(cmd *exec.Cmd, port int) {
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

func stopCmds(cmds ...*exec.Cmd) {
	for _, cmd := range cmds {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
