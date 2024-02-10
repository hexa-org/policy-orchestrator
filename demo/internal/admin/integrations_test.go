package admin_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/hexa-org/policy-orchestrator/demo/internal/admin"
	"github.com/hexa-org/policy-orchestrator/demo/internal/admin/test"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IntegrationsSuite struct {
	suite.Suite
	server *http.Server
	client *admin_test.MockClient
}

func TestIntegrations(t *testing.T) {
	suite.Run(t, new(IntegrationsSuite))
}

func (suite *IntegrationsSuite) SetupTest() {
	listener, _ := net.Listen("tcp", "localhost:0")
	suite.client = &admin_test.MockClient{Url: "http://noop"}
	suite.server = websupport.Create(
		listener.Addr().String(),
		admin.LoadHandlers("http://noop", suite.client),
		websupport.Options{})
	go websupport.Start(suite.server, listener)
	healthsupport.WaitForHealthy(suite.server)
}

func (suite *IntegrationsSuite) TearDownTest() {
	websupport.Stop(suite.server)
}

func (suite *IntegrationsSuite) TestListIntegrations() {
	resp := suite.must(http.Get(fmt.Sprintf("http://%s/integrations", suite.server.Addr)))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Discovery")
}

func (suite *IntegrationsSuite) TestListIntegrations_templateRenders() {
	resp := suite.must(http.Get(fmt.Sprintf("http://%s/integrations", suite.server.Addr)))
	body, _ := io.ReadAll(resp.Body)

	assert.Contains(suite.T(), string(body), "Google Cloud")
	assert.Contains(suite.T(), string(body), "Azure")
	assert.Contains(suite.T(), string(body), "Amazon Web Services")
	assert.Contains(suite.T(), string(body), "Open Policy Agent")
}

func (suite *IntegrationsSuite) TestListIntegrations_with_error() {
	suite.client.Errs = map[string]error{"http://noop/integrations": errors.New("oops")}

	resp, _ := http.Get(fmt.Sprintf("http://%s/integrations", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}

func (suite *IntegrationsSuite) TestNewIntegration_GoogleCloud() {
	resp := suite.must(http.Get(fmt.Sprintf("http://%s/integrations/new?provider=%s", suite.server.Addr, url.QueryEscape("google_cloud"))))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Install Cloud Provider")
}

func (suite *IntegrationsSuite) TestNewIntegration_Azure() {
	resp := suite.must(http.Get(fmt.Sprintf("http://%s/integrations/new?provider=%s", suite.server.Addr, url.QueryEscape("Azure"))))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Install Cloud Provider")
}

func (suite *IntegrationsSuite) TestNewIntegration_withoutQueryParam() {
	resp := suite.must(http.Get(fmt.Sprintf("http://%s/integrations/new", suite.server.Addr)))
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

func (suite *IntegrationsSuite) TestCreateIntegration_withGoogle() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	buf, contentType := suite.multipartFormSuccessGoogle()
	resp, _ := http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "google_cloud", suite.client.Provider)
	assert.Equal(suite.T(), "project:google-cloud-project-id", suite.client.Name)
}

func (suite *IntegrationsSuite) TestCreateIntegration_withAzure() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	buf, contentType := suite.multipartFormSuccessAzure()
	resp, _ := http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "azure", suite.client.Provider)
	assert.Equal(suite.T(), "tenant:aTenant", suite.client.Name)
}

func (suite *IntegrationsSuite) TestCreateIntegration_withAmazon() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	buf, contentType := suite.multipartFormSuccessAmazon()
	resp, _ := http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "amazon", suite.client.Provider)
	assert.Equal(suite.T(), "region:aRegion", suite.client.Name)
}

func (suite *IntegrationsSuite) TestCreateIntegration_withOPA() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	tests := []struct {
		name     string
		provider string
	}{
		{
			name:     "bundle server",
			provider: "bundle_server",
		},
		{
			name:     "Google Cloud Platform",
			provider: "gcp",
		},
		{
			name:     "AWS S3",
			provider: "aws",
		},
		{
			name:     "GITHUB",
			provider: "github",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			buf, contentType := suite.multipartFormSuccessOPA(tt.provider)
			resp, _ := http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf)
			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
			assert.Equal(suite.T(), "open_policy_agent", suite.client.Provider)
			assert.Equal(suite.T(), "opa-project-id:open-policy-agent", suite.client.Name)
		})
	}
}

func (suite *IntegrationsSuite) TestCreateIntegrationMissingKey_withAmazon() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	buf, contentType := suite.multipartFormMissingFileForAmazon()
	resp, _ := http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	all, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(all), "Something went wrong. Missing key file.")
}

func (suite *IntegrationsSuite) TestCreateIntegration_withError() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return(errors.New(""))
	buf, contentType := suite.multipartFormSuccessGoogle()
	resp := suite.must(http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf))
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	all, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(all), "Something went wrong. Unable to communicate with orchestrator.")
}

func (suite *IntegrationsSuite) TestCreateIntegration_withMissingKeyFile() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return(errors.New(""))
	buff, contentType := suite.multipartFormMissingFile()
	resp := suite.must(http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buff))
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	all, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(all), "Something went wrong. Missing key file.")
}

func (suite *IntegrationsSuite) TestCreateIntegration_withErroneousKeyFile() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	buf, contentType := suite.multipartFormErroneousFile()
	resp, _ := http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	all, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(all), "Something went wrong. unable to read key file, missing project")
}

func (suite *IntegrationsSuite) TestDeleteIntegration() {
	suite.client.On("DeleteIntegration", "http://noop/integrations/101").Return()
	resp := suite.must(http.Post(fmt.Sprintf("http://%s/integrations/101", suite.server.Addr), "", nil))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Discovery")
}

func (suite *IntegrationsSuite) TestDeleteIntegration_with_error() {
	suite.client.On("DeleteIntegration", "http://noop/integrations/101").Return(errors.New(""))
	resp := suite.must(http.Post(fmt.Sprintf("http://%s/integrations/101", suite.server.Addr), "", nil))
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (suite *IntegrationsSuite) must(resp *http.Response, _ error) *http.Response {
	return resp
}

func (suite *IntegrationsSuite) multipartFormMissingFile() (*bytes.Buffer, string) {
	return suite.multipartForm(func(writer *multipart.Writer) {
		provider, _ := writer.CreateFormField("provider")
		_, _ = provider.Write([]byte("google_cloud"))
	})
}

func (suite *IntegrationsSuite) multipartFormErroneousFile() (*bytes.Buffer, string) {
	return suite.multipartForm(func(writer *multipart.Writer) {
		provider, _ := writer.CreateFormField("provider")
		_, _ = provider.Write([]byte("google_cloud"))

		file, _ := writer.CreateFormFile("key", "aKey.json")
		_, _ = file.Write([]byte("{\"_____project_id\": \"google-cloud-project-id\"}"))
	})
}

func (suite *IntegrationsSuite) multipartFormSuccessGoogle() (*bytes.Buffer, string) {
	return suite.multipartForm(func(writer *multipart.Writer) {
		provider, _ := writer.CreateFormField("provider")
		_, _ = provider.Write([]byte("google_cloud"))

		file, _ := writer.CreateFormFile("key", "aKey.json")
		_, _ = file.Write([]byte("{\"type\": \"service_account\", \"project_id\": \"google-cloud-project-id\"}"))
	})
}

func (suite *IntegrationsSuite) multipartFormSuccessAzure() (*bytes.Buffer, string) {
	return suite.multipartForm(func(writer *multipart.Writer) {
		provider, _ := writer.CreateFormField("provider")
		_, _ = provider.Write([]byte("azure"))

		file, _ := writer.CreateFormFile("key", "aKey.json")
		_, _ = file.Write([]byte("{\"appId\": \"anAppId\", \"password\": \"aPassword\", \"tenant\": \"aTenant\"}"))
	})
}

func (suite *IntegrationsSuite) multipartFormSuccessAmazon() (*bytes.Buffer, string) {
	return suite.multipartForm(func(writer *multipart.Writer) {
		provider, _ := writer.CreateFormField("provider")
		_, _ = provider.Write([]byte("amazon"))

		file, _ := writer.CreateFormFile("key", "aKey.json")
		_, _ = file.Write([]byte("{\"region\": \"aRegion\"}"))
	})
}

func (suite *IntegrationsSuite) multipartFormSuccessOPA(bundleProvider string) (*bytes.Buffer, string) {
	return suite.multipartForm(func(writer *multipart.Writer) {
		provider, _ := writer.CreateFormField("provider")
		_, _ = provider.Write([]byte("open_policy_agent"))

		bundleFile := make(map[string]string)
		if bundleProvider == "bundle_server" {
			bundleFile["bundle_url"] = "http://opa-bundle-url"
		} else {
			bundleFile[bundleProvider] = "{}"
		}

		bundleFile["project_id"] = "opa-project-id"
		bundleBytes, _ := json.Marshal(bundleFile)
		file, _ := writer.CreateFormFile("key", "aKey.json")
		_, _ = file.Write(bundleBytes)
	})
}

func (suite *IntegrationsSuite) multipartFormMissingFileForAmazon() (*bytes.Buffer, string) {
	return suite.multipartForm(func(writer *multipart.Writer) {
		provider, _ := writer.CreateFormField("provider")
		_, _ = provider.Write([]byte("amazon"))
	})
}

func (suite *IntegrationsSuite) multipartForm(writeFormValues func(writer *multipart.Writer)) (*bytes.Buffer, string) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	writeFormValues(writer)
	contentType := writer.FormDataContentType()
	_ = writer.Close()
	return buf, contentType
}
