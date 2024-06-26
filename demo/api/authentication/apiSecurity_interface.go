package authentication

import (
	"io"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

type ClientInterface interface {
	Get(client HTTPClient, id string, key string, url string) (*http.Response, error)
	Post(client HTTPClient, id string, key string, url string, body io.Reader) (*http.Response, error)
}
