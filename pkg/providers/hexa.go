package providers

import (
	"bytes"
	"encoding/json"
	"io"
)

type Service interface {
	WritePolicies(policies []Policy, destination io.Writer) error
	ReadPolicies(source io.Reader) ([]Policy, error)
}

type Policy struct {
	Version string  `json:"version"`
	Action  string  `json:"action"`
	Object  Object  `json:"object"`
	Subject Subject `json:"subject"`
}

type Object struct {
	Resources []string `json:"resources"`
}

type Subject struct {
	AuthenticatedUsers []string `json:"authenticated_users"`
}

func Decode(policies []byte) ([]Policy, error) {
	var jsonResponse []Policy
	err := json.NewDecoder(bytes.NewReader(policies)).Decode(&jsonResponse)
	return jsonResponse, err
}

func Encode(policies []Policy) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(policies)
	return buffer.Bytes(), err
}
