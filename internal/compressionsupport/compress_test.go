package compressionsupport_test

import (
	"bytes"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarGzip(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources")
	tar, _ := compressionsupport.TarFromPath(join)

	var buffer bytes.Buffer
	gzipErr := compressionsupport.Gzip(&buffer, tar)
	assert.NoError(t, gzipErr)

	gzip, _ := compressionsupport.UnGzip(&buffer)
	dir := os.TempDir()

	unTarErr := compressionsupport.UnTarToPath(bytes.NewReader(gzip), dir)
	assert.NoError(t, unTarErr)
}

func TestTar(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/compressdir")
	_, err := compressionsupport.TarFromPath(join)
	assert.NoError(t, err)
}

func TestTar_withErr(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resourcez")
	_, err := compressionsupport.TarFromPath(join)
	assert.Error(t, err)
}

func TestUnTar(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/compressdir")
	tar, _ := compressionsupport.TarFromPath(join)

	dir := os.TempDir()
	err := compressionsupport.UnTarToPath(bytes.NewReader(tar), dir)
	assert.NoError(t, err)
}

func TestGzip(t *testing.T) {
	var buffer bytes.Buffer
	err := compressionsupport.Gzip(&buffer, []byte("someBytes"))
	assert.NoError(t, err)

	uncompressed, _ := compressionsupport.UnGzip(&buffer)
	assert.Equal(t, string(uncompressed), "someBytes")
}

func TestGzip_withErr(t *testing.T) {
	var buffer bytes.Buffer
	_, err := compressionsupport.UnGzip(&buffer)
	assert.Error(t, err)
}

func TestGzip_withCopyErr(t *testing.T) {
	var incorrect bytes.Buffer
	_ = compressionsupport.Gzip(&incorrect, []byte("someBytes"))
	incorrect.Write([]byte("oops"))
	_, err := compressionsupport.UnGzip(&incorrect)
	assert.Error(t, err)
}
