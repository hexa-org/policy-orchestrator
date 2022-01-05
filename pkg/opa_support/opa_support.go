package opa_support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type OpaSupport struct {
	client HTTPClient
	url string
}

func NewOpaSupport(client HTTPClient, url string) (*OpaSupport, error) {
	return &OpaSupport{client, url}, nil
}

type OpaQuery struct {
	Input map[string]interface{} `json:"input"`
}

type OpaResponse struct {
	Result bool
}

func (o *OpaSupport) Allow(input interface{}) (bool, error) {
	marshal, err := json.Marshal(input)
	if err != nil {
		return false, err
	}

	request, err := http.NewRequest("POST", o.url, bytes.NewBuffer(marshal))
	if err != nil {
		return false, err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response, err := o.client.Do(request)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("trouble closing a response body.")
		}
	}(response.Body)

	var jsonResponse OpaResponse
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		return false, err
	}

	if err != nil {
		return false, err
	}
	return jsonResponse.Result, nil
}

func OpaMiddleware(o *OpaSupport, next http.HandlerFunc, unauthorized http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := OpaQuery{map[string]interface{}{
			"method":     "GET",
			"path":       r.RequestURI,
			"principals": []interface{}{"allusers", "allauthenticatedusers", "sales@"},
		}}

		log.Println(fmt.Sprintf("Checking authorization for %v", input))

		allow, err := o.Allow(input)
		if err != nil {
			fmt.Println(err)
		}

		if !allow {
			unauthorized(w, r)
		} else {
			next(w, r)
		}
	}
}
