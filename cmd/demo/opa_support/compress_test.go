package opa_support_test

import (
	"github.com/stretchr/testify/assert"
	http2 "github.com/stretchr/testify/http"
	"hexa/cmd/demo/opa_support"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCompress(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	writer := http2.TestResponseWriter{}
	join := filepath.Join(file, "../../resources/bundles/bundle")
	opa_support.Compress(&writer, join)
	assert.NotEmpty(t, writer.Output)
}

func TestCompress_bad_path(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	writer := http2.TestResponseWriter{}
	join := filepath.Join(file, "../../resources/bundles/bundle_nope")
	opa_support.Compress(&writer, join)
	assert.Empty(t, writer.Output)
}
