package testhelper

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

type MockCognitoHttpClient struct {
	mock.Mock
}

func NewMockCognitoHttpClient() *MockCognitoHttpClient {
	return &MockCognitoHttpClient{}
}

func (m *MockCognitoHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// ExpectListUserPools - caller expects the listUserPools endpoint to be called
// with either err OR a valid response (see ListUserPoolsResponse())
func (m *MockCognitoHttpClient) ExpectListUserPools(err error) {
	var resp *http.Response
	if err == nil {
		resp = &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(ListUserPoolsResponse()))}
	}
	theFunc := mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodPost &&
			req.Header.Get("X-Amz-Target") == "AWSCognitoIdentityProviderService.ListUserPools"
	})

	m.On("Do", theFunc).Return(resp, err)
}

func (m *MockCognitoHttpClient) ExpectListResourceServers(err error) {
	var resp *http.Response
	if err == nil {
		expBytes, _ := json.Marshal(ListResourceServersOutput())
		resp = &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(expBytes))}
	}
	theFunc := mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == http.MethodPost &&
			req.Header.Get("X-Amz-Target") == "AWSCognitoIdentityProviderService.ListResourceServers"
	})

	m.On("Do", theFunc).Return(resp, err)
}
