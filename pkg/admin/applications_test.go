package admin_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ApplicationsSuite struct {
	suite.Suite
	server *http.Server
	client *admin_test.MockClient
}

func TestApplications(t *testing.T) {
	suite.Run(t, new(ApplicationsSuite))
}

func (suite *ApplicationsSuite) SetupTest() {
	listener, _ := net.Listen("tcp", "localhost:0")
	suite.client = &admin_test.MockClient{Url: "http://noop"}
	suite.server = websupport.Create(
		listener.Addr().String(),
		admin.LoadHandlers("http://noop", suite.client),
		websupport.Options{})
	go websupport.Start(suite.server, listener)
	healthsupport.WaitForHealthy(suite.server)
}

func (suite *ApplicationsSuite) TearDownTest() {
	websupport.Stop(suite.server)
}

func (suite *ApplicationsSuite) TestApplications_templateRenders() {
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)

	assert.Contains(suite.T(), string(body), "Provider")
	assert.Contains(suite.T(), string(body), "Platform Identifier")
	assert.Contains(suite.T(), string(body), "Name")
	assert.Contains(suite.T(), string(body), "Description")
}

func (suite *ApplicationsSuite) TestApplications() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription", ProviderName: "google_cloud"},
	}

	url := fmt.Sprintf("http://%s/applications", suite.server.Addr)
	resp, _ := http.Get(url)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Applications")
	assert.Contains(suite.T(), string(body), "anObjectId")
	assert.Contains(suite.T(), string(body), "aName")
	assert.Contains(suite.T(), string(body), "aDescription")
	assert.Contains(suite.T(), string(body), "Google Cloud Platform")
}

func (suite *ApplicationsSuite) TestApplications_with_error() {
	suite.client.Errs = map[string]error{"http://noop/applications": errors.New("oops")}

	resp, _ := http.Get(fmt.Sprintf("http://%s/applications", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}

func (suite *ApplicationsSuite) TestApplication() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anotherId", IntegrationId: "anotherIntegrationId", ObjectId: "anotherObjectId", Name: "anotherName", Description: "anotherDescription", ProviderName: "google_cloud"},
	}
	suite.client.DesiredPolicies = []admin.Policy{
		{admin.Meta{Version: "aVersion"}, []admin.Action{{"anAction"}}, admin.Subject{Members: []string{"aUser"}}, admin.Object{ResourceID: "aResourceId"}},
		{admin.Meta{Version: "anotherVersion"}, []admin.Action{{"anotherAction"}}, admin.Subject{Members: []string{"anotherUser"}}, admin.Object{ResourceID: "anotherResourceId"}},
	}

	identifier := "anotherId"
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/%s", suite.server.Addr, identifier))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Applications")
	assert.Contains(suite.T(), string(body), "anotherObjectId")
	assert.Contains(suite.T(), string(body), "anotherName")
	assert.Contains(suite.T(), string(body), "anotherDescription")

	assert.Contains(suite.T(), string(body), "aVersion")
	assert.Contains(suite.T(), string(body), "anAction")
	assert.Contains(suite.T(), string(body), "aUser")
	assert.Contains(suite.T(), string(body), "aResourceId")

	assert.Contains(suite.T(), string(body), "anotherVersion")
	assert.Contains(suite.T(), string(body), "anotherAction")
	assert.Contains(suite.T(), string(body), "anotherUser")
	assert.Contains(suite.T(), string(body), "anotherResourceId")
}

func (suite *ApplicationsSuite) TestApplication_withErroneousGet() {
	suite.client.Errs = map[string]error{"http://noop/applications/000": errors.New("oops")}

	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/000", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
	assert.Contains(suite.T(), string(body), "oops")
}

func (suite *ApplicationsSuite) TestApplication_withErroneousGetForPolicies() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription", ProviderName: "google_cloud"},
	}
	suite.client.Errs = map[string]error{
		"http://noop/applications/anId":          nil,
		"http://noop/applications/anId/policies": errors.New("oops"),
	}

	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/anId", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
	assert.Contains(suite.T(), string(body), "oops")
}

func (suite *ApplicationsSuite) TestApplication_Edit() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anotherId", IntegrationId: "anotherIntegrationId", ObjectId: "anotherObjectId", Name: "anotherName", Description: "anotherDescription", ProviderName: "google_cloud"},
	}
	suite.client.DesiredPolicies = []admin.Policy{
		{admin.Meta{Version: "aVersion"}, []admin.Action{{"anAction"}}, admin.Subject{Members: []string{"aUser"}}, admin.Object{ResourceID: "aResourceId"}},
		{admin.Meta{Version: "anotherVersion"}, []admin.Action{{"anotherAction"}}, admin.Subject{Members: []string{"anotherUser"}}, admin.Object{ResourceID: "anotherResourceId"}},
	}

	identifier := "anotherId"
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/%s/edit", suite.server.Addr, identifier))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Applications")
	assert.Contains(suite.T(), string(body), "anotherObjectId")
	assert.Contains(suite.T(), string(body), "anotherName")
	assert.Contains(suite.T(), string(body), "anotherDescription")

	assert.Contains(suite.T(), string(body), "action=\"/applications/anotherId\"")
}

func (suite *ApplicationsSuite) TestApplication_Edit_withErroneousGet() {
	suite.client.Errs = map[string]error{"http://noop/applications/anId": errors.New("oops")}
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/anId/edit", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
	assert.Contains(suite.T(), string(body), "oops")
}

func (suite *ApplicationsSuite) TestApplication_Edit_withErroneousGetForPolicies() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription", ProviderName: "google_cloud"},
	}
	suite.client.Errs = map[string]error{"http://noop/applications/anId": nil,
		"http://noop/applications/anId/policies": errors.New("shoot"),
	}

	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/anId/edit", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
	assert.Contains(suite.T(), string(body), "shoot")
}

func (suite *ApplicationsSuite) TestApplication_Update() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription", ProviderName: "google_cloud"},
	}
	suite.client.DesiredPolicies = []admin.Policy{
		{admin.Meta{Version: "aVersion"}, []admin.Action{{"anAction"}}, admin.Subject{Members: []string{"aUser"}}, admin.Object{ResourceID: "aResourceId"}},
	}

	identifier := "anId"
	resp, _ := http.Post(fmt.Sprintf("http://%s/applications/%s", suite.server.Addr, identifier), "application/json", nil)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Applications")
	assert.Contains(suite.T(), string(body), "anObjectId")
	assert.Contains(suite.T(), string(body), "aName")
	assert.Contains(suite.T(), string(body), "aDescription")

	assert.Contains(suite.T(), string(body), "aVersion")
	assert.Contains(suite.T(), string(body), "anAction")
	assert.Contains(suite.T(), string(body), "aUser")
	assert.Contains(suite.T(), string(body), "aResourceId")
}

func (suite *ApplicationsSuite) TestApplication_Update_withErroneousGet() {
	suite.client.Errs = map[string]error{"http://noop/applications/anId": errors.New("oops")}

	resp, _ := http.Post(fmt.Sprintf("http://%s/applications/anId", suite.server.Addr), "application/json", nil)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
	assert.Contains(suite.T(), string(body), "oops")
}

func (suite *ApplicationsSuite) TestApplication_Update_withErroneousGetForPolicies() {
	suite.client.Errs = map[string]error{
		"http://noop/applications/anId/policies": errors.New("shoot"),
		"http://noop/applications/anId":          nil,
	}

	resp, _ := http.Post(fmt.Sprintf("http://%s/applications/anId", suite.server.Addr), "application/json", nil)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
	assert.Contains(suite.T(), string(body), "shoot")
}

func (suite *ApplicationsSuite) TestApplication_Policies() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription", ProviderName: "google_cloud"},
	}
	suite.client.DesiredPolicies = []admin.Policy{
		{admin.Meta{Version: "aVersion"}, []admin.Action{{"anAction"}}, admin.Subject{Members: []string{"aUser"}}, admin.Object{ResourceID: "aResourceId"}},
	}

	identifier := "anId"
	resp, _ := http.Get(fmt.Sprintf("http://%s/applications/%s/policies", suite.server.Addr, identifier))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "{\n  \"policies\": []\n}")
}
