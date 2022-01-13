package providers_test

import (
	"bytes"
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/pkg/providers"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEncoding(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	jsonFile, _ := os.Open(filepath.Join(file, "../test/policy.json"))
	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	policies, _ := providers.Decode(jsonBytes)
	assert.Equal(t, 4, len(policies))

	policiesBytes, _ := providers.Encode([]providers.Policy{policies[0]})

	expected := `
[
  {
    "version": "0.1",
    "action": "GET",
    "object": {
      "resources": [
        "/"
      ]
    },
    "subject": {
      "authenticated_users": [
        "allusers"
      ]
    }
  }
]
`
	expectedCompact := new(bytes.Buffer)
	_ = json.Compact(expectedCompact, []byte(expected))
	actualCompact := new(bytes.Buffer)
	_ = json.Compact(actualCompact, policiesBytes)
	assert.Equal(t, expectedCompact.String(), actualCompact.String())
}
