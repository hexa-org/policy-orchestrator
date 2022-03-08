package gzipsupport_test

import (
	"github.com/hexa-org/policy-orchestrator/cmd/democonfig/gzipsupport"
	"github.com/stretchr/testify/assert"
	http2 "github.com/stretchr/testify/http"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCompress(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	writer := http2.TestResponseWriter{}
	join := filepath.Join(file, "../../resources/bundles/bundle")
	gzipsupport.Compress(&writer, join)
	assert.NotEmpty(t, writer.Output)
}

func TestCompress_bad_path(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	writer := http2.TestResponseWriter{}
	join := filepath.Join(file, "../../resources/bundles/bundle_nope")
	gzipsupport.Compress(&writer, join)
	assert.Empty(t, writer.Output)
}
