package decisionsupportproviders

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type OpaDecisionProvider struct {
	Client     HTTPClient
	Url        string
	Principals []interface{}
}

type OpaQuery struct {
	Input map[string]interface{} `json:"input"`
}

func (o OpaDecisionProvider) BuildInput(r *http.Request) (any interface{}, err error) {
	return OpaQuery{map[string]interface{}{
		"method":     "GET",
		"path":       r.URL.Path,
		"principals": o.Principals,
	}}, nil
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type OpaResponse struct {
	Result bool
}

func (o OpaDecisionProvider) Allow(any interface{}) (bool, error) {
	marshal, _ := json.Marshal(any.(OpaQuery))
	request, _ := http.NewRequest("POST", o.Url, bytes.NewBuffer(marshal))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response, err := o.Client.Do(request)
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
