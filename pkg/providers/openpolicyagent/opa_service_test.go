package openpolicyagent_test

import (
	"bytes"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/openpolicyagent"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	policies := []provider.PolicyInfo{
		{Version: "0.1", Action: "GET", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"allusers"}}, Object: provider.ObjectInfo{Resources: []string{"/"}}},
		{Version: "0.1", Action: "GET", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"sales@", "marketing@"}}, Object: provider.ObjectInfo{Resources: []string{"/sales", "/marketing"}}},
		{Version: "0.1", Action: "GET", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"accounting@"}}, Object: provider.ObjectInfo{Resources: []string{"/accounting"}}},
		{Version: "0.1", Action: "GET", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"humanresources@"}}, Object: provider.ObjectInfo{Resources: []string{"/humanresources"}}},
	}

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := openpolicyagent.OpaService{ResourcesDirectory: resourcesDirectory}

	actualRegoBytes := new(strings.Builder)
	_ = service.WritePolicies(policies, actualRegoBytes)

	regoFile, _ := os.Open(filepath.Join(resourcesDirectory, "./bundles/bundle/policy.rego"))
	regoBytes, _ := ioutil.ReadAll(regoFile)

	assert.Equal(t, string(regoBytes), actualRegoBytes.String())
}

func TestWriteEmpty(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := openpolicyagent.OpaService{ResourcesDirectory: resourcesDirectory}

	actualRegoBytes := new(strings.Builder)
	_ = service.WritePolicies([]provider.PolicyInfo{}, actualRegoBytes)

	assert.Equal(t, "package authz\nimport future.keywords.in\ndefault allow = false", actualRegoBytes.String())
}

func TestRead(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := openpolicyagent.OpaService{ResourcesDirectory: resourcesDirectory}

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
	service := openpolicyagent.OpaService{ResourcesDirectory: resourcesDirectory}

	reader := bytes.NewReader([]byte(""))
	_, err := service.ReadPolicies(reader)
	assert.Contains(t, err.Error(), "unexpected token")
}

func TestReadWrite(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	service := openpolicyagent.OpaService{ResourcesDirectory: resourcesDirectory}

	regoFile, _ := os.Open(filepath.Join(resourcesDirectory, "./bundles/bundle/policy.rego"))
	regoBytes, _ := ioutil.ReadAll(regoFile)
	policies, _ := service.ReadPolicies(bytes.NewReader(regoBytes))

	actualRegoBytes := new(strings.Builder)
	_ = service.WritePolicies(policies, actualRegoBytes)

	assert.Equal(t, string(regoBytes), actualRegoBytes.String())
}
