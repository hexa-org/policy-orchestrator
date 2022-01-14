package open_policy_agent_test

import (
	"bytes"
	"github.com/hexa-org/policy-orchestrator/pkg/providers"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/open_policy_agent"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	jsonFile, _ := os.Open(filepath.Join(file, "../../test/policy.json"))
	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	policies, _ := providers.Decode(jsonBytes)

	resourcesDirectory := filepath.Join(file, "../resources")
	service := open_policy_agent.NewOpaService(resourcesDirectory)

	actualRegoBytes := new(strings.Builder)
	_ = service.WritePolicies(policies, actualRegoBytes)

	regoFile, _ := os.Open(filepath.Join(resourcesDirectory, "./bundles/bundle/policy.rego"))
	regoBytes, _ := ioutil.ReadAll(regoFile)

	assert.Equal(t, string(regoBytes), actualRegoBytes.String())
}

func TestWriteEmpty(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := open_policy_agent.NewOpaService(resourcesDirectory)

	actualRegoBytes := new(strings.Builder)
	_ = service.WritePolicies([]providers.Policy{}, actualRegoBytes)

	assert.Equal(t, "package authz\nimport future.keywords.in\ndefault allow = false", actualRegoBytes.String())
}

func TestRead(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := open_policy_agent.NewOpaService(resourcesDirectory)

	regoFile, _ := os.Open(filepath.Join(resourcesDirectory, "./bundles/bundle/policy.rego"))
	regoBytes, _ := ioutil.ReadAll(regoFile)
	reader := bytes.NewReader(regoBytes)

	policies, _ := service.ReadPolicies(reader)
	assert.Equal(t, 4, len(policies))
	assert.Equal(t, "GET", policies[0].Action)
	assert.Equal(t, []string{"/"}, policies[0].Object.Resources)
	assert.Equal(t, []string{"allusers"}, policies[0].Subject.AuthenticatedUsers)
}

func TestRead_failed(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := open_policy_agent.NewOpaService(resourcesDirectory)

	reader := bytes.NewReader([]byte(""))
	_, err := service.ReadPolicies(reader)
	assert.Contains(t, err.Error(), "unexpected token")
}

func TestReadWrite(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := open_policy_agent.NewOpaService(resourcesDirectory)

	regoFile, _ := os.Open(filepath.Join(resourcesDirectory, "./bundles/bundle/policy.rego"))
	regoBytes, _ := ioutil.ReadAll(regoFile)
	policies, _ := service.ReadPolicies(bytes.NewReader(regoBytes))

	actualRegoBytes := new(strings.Builder)
	_ = service.WritePolicies(policies, actualRegoBytes)

	assert.Equal(t, string(regoBytes), actualRegoBytes.String())
}
