package decisionsupportproviders_test

import (
	"bytes"
	"errors"
	"github.com/hexa-org/policy-orchestrator/internal/decisionsupportproviders"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOpaDecisionProvider_BuildInput(t *testing.T) {
	provider := decisionsupportproviders.OpaDecisionProvider{
		Principal: "sales@hexaindustries.io",
	}

	req, _ := http.NewRequest("POST", "https://aDomain.com/noop", nil)
	req.RequestURI = "/noop"
	query, _ := provider.BuildInput(req)
	casted := query.(decisionsupportproviders.OpaQuery).Input
	assert.Equal(t, "https:POST:/noop", casted["method"])
	assert.Equal(t, "sales@hexaindustries.io", casted["principal"])
}

func TestOpaDecisionProvider_BuildInput_RemovesQueryParams(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":true}")
	provider := decisionsupportproviders.OpaDecisionProvider{Client: mockClient, Url: "aUrl"}

	req, _ := http.NewRequest("GET", "http://aDomain.com/noop/?param=aParam", nil)
	req.RequestURI = "/noop"
	query, _ := provider.BuildInput(req)

	assert.Equal(t, "http:GET:/noop", query.(decisionsupportproviders.OpaQuery).Input["method"])
}

type MockClient struct {
	mock.Mock
	response []byte
	err      error
}

func (m *MockClient) Do(_ *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: 200, Body: r}, m.err
}

func TestOpaDecisionProvider_Allow(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":true}")
	provider := decisionsupportproviders.OpaDecisionProvider{Client: mockClient, Url: "aUrl"}

	req, _ := http.NewRequest("GET", "http://aDomain.com/noop", nil)
	req.RequestURI = "/noop"
	query, _ := provider.BuildInput(req)

	allow, _ := provider.Allow(query)
	assert.Equal(t, true, allow)
}

func TestOpaDecisionProvider_AllowWithRequestErr(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":true}")
	mockClient.err = errors.New("oops")
	provider := decisionsupportproviders.OpaDecisionProvider{Client: mockClient, Url: "aUrl"}

	req, _ := http.NewRequest("GET", "http://aDomain.com/noop", nil)
	req.RequestURI = "/noop"
	query, _ := provider.BuildInput(req)

	allow, err := provider.Allow(query)
	assert.Equal(t, "oops", err.Error())
	assert.Equal(t, false, allow)
}

func TestOpaDecisionProvider_AllowWithResponseErr(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("__bad__ {\"result\":true}")
	provider := decisionsupportproviders.OpaDecisionProvider{Client: mockClient, Url: "aUrl"}

	req, _ := http.NewRequest("GET", "http://aDomain.com/noop", nil)
	req.RequestURI = "/noop"
	query, _ := provider.BuildInput(req)

	allow, err := provider.Allow(query)
	assert.Equal(t, "invalid character '_' looking for beginning of value", err.Error())
	assert.Equal(t, false, allow)
}
