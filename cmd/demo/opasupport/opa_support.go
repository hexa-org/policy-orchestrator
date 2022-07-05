package opasupport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type OpaSupport struct {
	client       HTTPClient
	url          string
	unauthorized http.HandlerFunc
	skip         []string
}

func NewOpaSupport(client HTTPClient, url string, unauthorized http.HandlerFunc) *OpaSupport {
	return &OpaSupport{client, url, unauthorized,
		[]string{"/health", "/metrics", "/styles", "/images", "/bundle"}}
}

type OpaQuery struct {
	Input map[string]interface{} `json:"input"`
}

type OpaResponse struct {
	Result bool
}

func (o *OpaSupport) Allow(input interface{}) (bool, error) {
	marshal, _ := json.Marshal(input)
	request, _ := http.NewRequest("POST", o.url, bytes.NewBuffer(marshal))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response, err := o.client.Do(request)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	var jsonResponse OpaResponse
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		return false, err
	}
	return jsonResponse.Result, nil
}

func (o *OpaSupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, s := range o.skip {
			if strings.HasPrefix(r.RequestURI, s) {
				next.ServeHTTP(w, r)
				return
			}
		}
		input := OpaQuery{map[string]interface{}{
			"method":     "http:GET",
			"path":       strings.Split(r.RequestURI, "?")[0],
			"principal": "sales@hexaindustries.io",
		}}
		log.Println(fmt.Sprintf("Checking authorization for %v", input))

		allow, err := o.Allow(input)
		if !allow || err != nil {
			o.unauthorized(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
