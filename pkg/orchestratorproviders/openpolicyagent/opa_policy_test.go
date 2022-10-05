package openpolicyagent_test

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupportproviders"
	assert "github.com/stretchr/testify/require"
)

func TestPolicy(t *testing.T) {
	openPolicyAgent := exec.Command("opa", "run", "--server", "--addr", ":8887")
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
	dataDo, err := http.DefaultClient.Do(dataReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, dataDo.StatusCode)

	rego, err := os.ReadFile(filepath.Join(file, "../resources/bundles/bundle/policy.rego"))
	assert.NoError(t, err)
	regoReq, err := http.NewRequest(http.MethodPut, "http://localhost:8887/v1/policies/authz", bytes.NewBuffer(rego))
	assert.NoError(t, err)
	regoDo, err := http.DefaultClient.Do(regoReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, regoDo.StatusCode)

	provider := decisionsupportproviders.OpaDecisionProvider{
		Client: http.DefaultClient,
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
	log.Printf("Starting command: %v\n", cmd)
	errCh := make(chan error)

	go func() {
		err := cmd.Run()
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case <-time.After(time.Second):
		log.Println("Command started")
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Unable to start cmd %v\n", err)
		}
	}
}
