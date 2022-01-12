package admin_test

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"hexa/pkg/admin"
	"hexa/pkg/admin/test"
	"hexa/pkg/web_support"
	"io"
	"mime/multipart"
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

	suite.client = new(admin_test.MockClient)
	suite.server = web_support.Create(
		"localhost:8883",
		admin.LoadHandlers("http://noop", suite.client),
		web_support.Options{ResourceDirectory: resourcesDirectory})
	go web_support.Start(suite.server)
	web_support.WaitForHealthy(suite.server)
}

func (suite *IntegrationsSuite) TearDownTest() {
	web_support.Stop(suite.server)
}

///

func (suite *IntegrationsSuite) TestListIntegrations() {
	resp := suite.must(http.Get("http://localhost:8883/integrations"))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Discovery")
}

func (suite *IntegrationsSuite) TestNewIntegration() {
	resp := suite.must(http.Get("http://localhost:8883/integrations/new"))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Install Cloud Provider")
}

func (suite *IntegrationsSuite) TestCreateIntegration() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return()
	buf, contentType := suite.multipartForm()
	resp := suite.must(http.Post("http://localhost:8883/integrations", contentType, buf))
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *IntegrationsSuite) TestCreateIntegration_with_error() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return(errors.New(""))
	buf, contentType := suite.multipartForm()
	resp := suite.must(http.Post("http://localhost:8883/integrations", contentType, buf))
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)
}

func (suite *IntegrationsSuite) TestCreateIntegration_missing_key_file() {
	suite.client.On("CreateIntegration", "http://noop/integrations").Return(errors.New(""))
	buff, contentType := suite.multipartFormMissingFile()
	resp := suite.must(http.Post("http://localhost:8883/integrations", contentType, buff))
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	all, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(all), "Something went wrong. Missing key file.")
}

func (suite *IntegrationsSuite) TestDeleteIntegration() {
	suite.client.On("DeleteIntegration", "http://noop/integrations/101").Return()
	resp := suite.must(http.Post("http://localhost:8883/integrations/101", "", nil))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Discovery")
}

func (suite *IntegrationsSuite) TestDeleteIntegration_with_error() {
	suite.client.On("DeleteIntegration", "http://noop/integrations/101").Return(errors.New(""))
	resp := suite.must(http.Post("http://localhost:8883/integrations/101", "", nil))
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)
}

///

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
