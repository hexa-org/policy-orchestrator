package hexapolicysupport_test

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicysupport"
)

func TestReadFile(t *testing.T) {
	idqlPath := getFile()

	policies, err := hexapolicysupport.ParsePolicyFile(idqlPath)
	assert.NoError(t, err, "File %s not parsed", idqlPath)

	assert.Equal(t, 4, len(policies), "Expecting 4 policies")
}

func TestWriteFile(t *testing.T) {
	policies, err := hexapolicysupport.ParsePolicyFile(getFile())
	assert.NoError(t, err, "File %s not parsed", getFile())

	rand.Seed(time.Now().UnixNano())
	dir := t.TempDir()

	tmpFile := filepath.Join(dir, fmt.Sprintf("idqldata-%d.json", rand.Uint64()))
	err = hexapolicysupport.WritePolicies(tmpFile, policies)
	assert.NoError(t, err, "Check error on writing policy")

	policyCopy, err := hexapolicysupport.ParsePolicyFile(tmpFile)
	assert.Equal(t, 4, len(policyCopy), "4 policies in copy parsed")
	assert.Equal(t, policies, policyCopy, "Check that the copy is the same as the original")
}

func getFile() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(file, "../test/data.json")
}
