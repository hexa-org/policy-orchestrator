package demosmoke_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDemoFlow(t *testing.T) {

	demo := makeGoCommand("/cmd/demo/demo.go", []string{"PORT=8886", "OPA_SERVER_URL: http://localhost:8887/v1/data/authz/allow"}, []string{})
	demoConfig := makeGoCommand("/cmd/democonfig/democonfig.go", []string{"PORT=8889"}, []string{})
	demoProxy := makeGoCommand("/cmd/demoproxy/demoproxy.go", []string{"PORT=8890", "REMOTER_URL=http://localhost:8886"}, []string{""})

	_, file, _, _ := runtime.Caller(0)
	config := filepath.Join(file, "../../../deployments/opa-server/config/config.yaml")
	openPolicyAgent := exec.Command("opa", "run", "--server", "--addr", ":8887", "-c", config)
	openPolicyAgent.Env = os.Environ()
	openPolicyAgent.Env = append(openPolicyAgent.Env, "HEXA_DEMO_URL=http://localhost:8889")

	start(demo, 8886)
	start(demoConfig, 8889)
	start(demoProxy, 8890)
	start(openPolicyAgent, 8887)

	resp, _ := http.Get(fmt.Sprintf("http://%s/", fmt.Sprintf("localhost:%v", 8890)))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Great news, you're able to access this page.")

	resp, _ = http.Get(fmt.Sprintf("http://%s/accounting", fmt.Sprintf("localhost:%v", 8890)))
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Sorry, you're not able to access this page.")

	resp, _ = http.Get(fmt.Sprintf("http://%s/humanresources", fmt.Sprintf("localhost:%v", 8890)))
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Sorry, you're not able to access this page.")

	stopAll(openPolicyAgent, demoProxy, demoConfig, demo)
}

func makeGoCommand(cmdString string, envs []string, args []string) *exec.Cmd {
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(file, "../../../")
	commandPath := filepath.Join(path, cmdString)

	args = append([]string{commandPath}, args...)
	args = append([]string{"run"}, args...)

	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, envs...)

	return cmd
}

func start(cmd *exec.Cmd, port int) {
	log.Printf("starting cmd %v\n", cmd)
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	go func() {
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Println(stdErr.String())
			log.Println("unable to makeGoCommand app.")
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

func stopAll(cmds ...*exec.Cmd) {
	for _, cmd := range cmds {
		err := cmd.Process.Kill()
		if err != nil {
			log.Println("shoot, lost the process.")
		}
	}
}
