package google_cloud_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

func Resource(name string) []byte {
	_, f, _, _ := runtime.Caller(0)
	json := filepath.Join(f, fmt.Sprintf("./../%v", name))
	content, _ := ioutil.ReadFile(json)
	return content
}
