package admin_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
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
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")

	listener, _ := net.Listen("tcp", "localhost:0")
	suite.client = new(admin_test.MockClient)
	suite.server = websupport.Create(
		listener.Addr().String(),
		admin.LoadHandlers("http://noop", suite.client),
		websupport.Options{ResourceDirectory: resourcesDirectory})
	go websupport.Start(suite.server, listener)
	websupport.WaitForHealthy(suite.server)
}

func (suite *IntegrationsSuite) TearDownTest() {
	websupport.Stop(suite.server)
}

func (suite *IntegrationsSuite) TestListIntegrations() {
	resp := suite.must(http.Get(fmt.Sprintf("http://%s/integrations", suite.server.Addr)))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Discovery")
}

func (suite *IntegrationsSuite) TestListIntegrations_with_error() {
	suite.client.Err = errors.New("oops")
	resp, _ := http.Get(fmt.Sprintf("http://%s/integrations", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}

func (suite *IntegrationsSuite) TestNewIntegration() {
	resp := suite.must(http.Get(fmt.Sprintf("http://%s/integrations/new", suite.server.Addr)))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Install Cloud Provider")
}

func (suite *IntegrationsSuite) TestCreateIntegration() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	buf, contentType := suite.multipartForm()
	resp := suite.must(http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf))
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *IntegrationsSuite) TestCreateIntegration_with_error() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return(errors.New(""))
	buf, contentType := suite.multipartForm()
	resp := suite.must(http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buf))
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (suite *IntegrationsSuite) TestCreateIntegration_missing_key_file() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return(errors.New(""))
	buff, contentType := suite.multipartFormMissingFile()
	resp := suite.must(http.Post(fmt.Sprintf("http://%s/integrations", suite.server.Addr), contentType, buff))
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	all, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(all), "Something went wrong. Missing key file.")
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
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	provider, _ := writer.CreateFormField("provider")
	_, _ = provider.Write([]byte("google cloud"))
	contentType := writer.FormDataContentType()
	_ = writer.Close()
	return buf, contentType
}

func (suite *IntegrationsSuite) multipartForm() (*bytes.Buffer, string) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	provider, _ := writer.CreateFormField("provider")
	_, _ = provider.Write([]byte("google cloud"))
	file, _ := writer.CreateFormFile("key", "aKey.json")
	_, _ = file.Write([]byte("aKey"))
	contentType := writer.FormDataContentType()
	_ = writer.Close()
	return buf, contentType
}
