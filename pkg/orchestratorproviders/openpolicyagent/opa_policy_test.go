package openpolicyagent_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupportproviders"
	"github.com/stretchr/testify/assert"
)

func TestPolicy(t *testing.T) {
	openPolicyAgent := exec.Command("opa", "run", "--server", "--addr", "localhost:8887")
	openPolicyAgent.Stdout = os.Stdout
	openPolicyAgent.Stderr = os.Stderr
	openPolicyAgent.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	startCmd(openPolicyAgent, 8887)
	defer func() {
		_ = syscall.Kill(-openPolicyAgent.Process.Pid, syscall.SIGKILL)
	}()

	_, file, _, _ := runtime.Caller(0)

	data, _ := os.ReadFile(filepath.Join(file, "../resources/bundles/bundle/data.json"))
	dataReq, _ := http.NewRequest(http.MethodPut, "http://localhost:8887/v1/data/bundle", bytes.NewBuffer(data))
	dataDo, _ := (&http.Client{}).Do(dataReq)
	assert.Equal(t, http.StatusNoContent, dataDo.StatusCode)

	rego, _ := os.ReadFile(filepath.Join(file, "../resources/bundles/bundle/policy.rego"))
	regoReq, _ := http.NewRequest(http.MethodPut, "http://localhost:8887/v1/policies/authz", bytes.NewBuffer(rego))
	regoDo, _ := (&http.Client{}).Do(regoReq)
	assert.Equal(t, http.StatusOK, regoDo.StatusCode)

	provider := decisionsupportproviders.OpaDecisionProvider{
		Client: &http.Client{},
		Url:    "http://localhost:8887/v1/data/authz/allow",
	}

	shouldAllow(t, provider, "http:GET:/", "")
	shouldAllow(t, provider, "http:GET:/", "any@google.com")
	shouldAllow(t, provider, "http:GET:/", "sales@hexaindustries.io")

	shouldNotAllow(t, provider, "http:GET:/sales", "")
	shouldAllow(t, provider, "http:GET:/sales", "any@google.com")
	shouldAllow(t, provider, "http:GET:/sales", "sales@hexaindustries.io")
	shouldAllow(t, provider, "http:GET:/sales", "marketing@hexaindustries.io")

	shouldNotAllow(t, provider, "http:GET:/marketing", "")
	shouldAllow(t, provider, "http:GET:/marketing", "any@google.com")
	shouldAllow(t, provider, "http:GET:/marketing", "sales@hexaindustries.io")
	shouldAllow(t, provider, "http:GET:/marketing", "marketing@hexaindustries.io")

	shouldNotAllow(t, provider, "http:GET:/accounting", "")
	shouldNotAllow(t, provider, "http:GET:/accounting", "any@google.com")
	shouldNotAllow(t, provider, "http:GET:/accounting", "sales@hexaindustries.io")
	shouldAllow(t, provider, "http:GET:/accounting", "accounting@hexaindustries.io")

	shouldNotAllow(t, provider, "http:GET:/humanresources", "")
	shouldNotAllow(t, provider, "http:GET:/humanresources", "any@google.com")
	shouldNotAllow(t, provider, "http:GET:/humanresources", "sales@hexaindustries.io")
	shouldAllow(t, provider, "http:GET:/humanresources", "humanresources@hexaindustries.io")
}

func shouldAllow(t *testing.T, provider decisionsupportproviders.OpaDecisionProvider, action string, principal string) {
	assert.True(t, allows(provider, action, principal))
}

func shouldNotAllow(t *testing.T, provider decisionsupportproviders.OpaDecisionProvider, action string, principal string) {
	assert.False(t, allows(provider, action, principal))
}

func allows(provider decisionsupportproviders.OpaDecisionProvider, action string, principal string) bool {
	allow, _ := provider.Allow(decisionsupportproviders.OpaQuery{Input: map[string]interface{}{"method": action, "principal": principal}})
	return allow
}

func startCmd(cmd *exec.Cmd, port int) {
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
			isLive = true
		}
	}
}
