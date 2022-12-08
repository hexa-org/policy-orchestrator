package testsupport

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData interface {
	SetUp()
	TearDown()
}

func WithSetUp[T TestData](data T, test func(data T)) {
	data.SetUp()
	test(data)
	data.TearDown()
}

func AssertGetContains(t *testing.T, app *http.Server, uri string, expected string) {
	version, _ := http.Get(fmt.Sprintf("http://%s%s", app.Addr, uri))
	assert.Equal(t, http.StatusOK, version.StatusCode)
	versionBody, _ := io.ReadAll(version.Body)
	assert.Contains(t, string(versionBody), expected)
}

func AssertGetWithCookieContains(t *testing.T, app *http.Server, client *http.Client, cookie *http.Cookie, uri string, expected string) {
	request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s%s", app.Addr, uri), nil)
	request.AddCookie(cookie)
	response, _ := client.Do(request)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(body), expected)
}

func AssertExists(file []byte, err error) []byte {
	if err != nil {
		panic("unable to read file.")
	}
	return file
}
