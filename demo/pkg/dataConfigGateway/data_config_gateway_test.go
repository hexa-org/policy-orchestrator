package dataConfigGateway

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	suite.Suite
	testDir          string
	Data             *ConfigData
	integDataGateway IntegrationsDataGateway
	appDataGateway   ApplicationsDataGateway
	ConfigPath       string
	test1Id          string
	testApp          ApplicationRecord
	mu               sync.Mutex
}

func (suite *testSuite) SetupTest() {
	suite.mu.Lock()
}

func (suite *testSuite) TearDownTest() {
	suite.mu.Unlock()
}

func TestConfigGateway(t *testing.T) {
	fmt.Println("Running Config Data Gateway Tests...")
	s := testSuite{}

	var err error
	_ = os.Setenv(sdk.EnvTestProvider, sdk.ProviderTypeMock)
	dir, err := os.MkdirTemp("", "hexa-orchestrator-*")
	assert.NoError(t, err, "Error creating temp dir")
	s.testDir = dir

	testConfigPath := filepath.Join(s.testDir, ".hexa", "config.json")
	s.ConfigPath = testConfigPath
	_ = os.Setenv(EnvIntegrationConfigFile, testConfigPath)

	s.Data, err = NewIntegrationConfigData()
	s.integDataGateway = s.Data
	s.appDataGateway = s.Data.GetApplicationDataGateway()

	assert.NoError(t, err, "Should be no error opening new config")
	assert.Equal(t, testConfigPath, s.Data.ConfigFile, "Config file should be equal")

	if err == nil {
		suite.Run(t, &s)
	}

	_ = os.RemoveAll(s.testDir)
	fmt.Println("** Test complete **")
}

func (s *testSuite) Test1_IG_AddIntegration() {
	_, file, _, _ := runtime.Caller(0)
	awsTestFile := filepath.Join(file, "../test/aws_test.json")
	keyfile, err := os.ReadFile(awsTestFile)
	assert.NoError(s.T(), err, "check aws file read")
	id, err := s.integDataGateway.Create("", sdk.ProviderTypeAwsApiGW, keyfile)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), id)
	s.test1Id = id
}

func (s *testSuite) Test1_IG_FindIntegration() {
	recs := s.integDataGateway.Find()

	assert.Len(s.T(), recs, 1)
	assert.Equal(s.T(), s.test1Id, recs[0].ID, "Should be equal")

}

func (s *testSuite) Test2_IG_FindIntegrationById() {
	record, err := s.integDataGateway.FindById(s.test1Id)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), s.test1Id, record.ID, "Should be equal")

	_, err = s.integDataGateway.FindById("notfound")
	assert.Error(s.T(), err, "integration does not exist")

	sdkInt := s.Data.GetIntegration(s.test1Id)
	assert.Equal(s.T(), sdkInt.Alias, record.ID)
}

func (s *testSuite) Test3_IG_Delete() {
	_, file, _, _ := runtime.Caller(0)
	azTestFile := filepath.Join(file, "../test/azure_test.json")
	keyfile, err := os.ReadFile(azTestFile)
	assert.NoError(s.T(), err)
	id, err := s.integDataGateway.Create("", sdk.ProviderTypeAzure, keyfile)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), s.test1Id, id)

	recs := s.integDataGateway.Find()
	assert.Len(s.T(), recs, 2)

	err = s.integDataGateway.Delete(id)
	assert.NoError(s.T(), err)

	assert.Len(s.T(), s.Data.Integrations, 1, "Should only be the original")

	err = s.integDataGateway.Delete("notfound")
	assert.Error(s.T(), err, "integration does not exist")
	// Check to make sure nothing deleted
	assert.Len(s.T(), s.Data.Integrations, 1, "Should only be the original")
}

func (s *testSuite) Test4_AG_Find() {
	// Gen more test data
	_, file, _, _ := runtime.Caller(0)
	azTestFile := filepath.Join(file, "../test/azure_test.json")
	keyfile, err := os.ReadFile(azTestFile)
	id2, err := s.integDataGateway.Create("", sdk.ProviderTypeAzure, keyfile)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), s.test1Id, id2)

	apps, err := s.appDataGateway.Find(true)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), apps, 2)

	s.testApp = apps[0]

	// reset for next text (because mock app has same objectid)
	err = s.Data.Delete(id2)
	assert.NoError(s.T(), err)
}

func (s *testSuite) Test5_AG_FindByObjectId() {

	apps, err := s.appDataGateway.Find(true)
	assert.NoError(s.T(), err)
	assert.Greater(s.T(), len(apps), 0, "Should not be empty")
	testObj := apps[0].ObjectId

	app, err := s.appDataGateway.FindByObjectId(testObj)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), apps[0].IntegrationId, app.IntegrationId)

	_, err = s.appDataGateway.FindByObjectId("dummy")
	assert.Error(s.T(), err, "application dummy not found")
}

func (s *testSuite) Test6_AG_FindById() {
	testId := s.testApp.ID
	app, err := s.appDataGateway.FindById(testId)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), s.testApp.IntegrationId, app.IntegrationId)

	_, err = s.appDataGateway.FindById("dummy")
	assert.Error(s.T(), err, "application dummy not found")
}

func (s *testSuite) Test7_AG_DeleteById() {
	err := s.appDataGateway.DeleteById(s.testApp.ID)
	assert.NoError(s.T(), err)
	integration := s.Data.Integrations[s.test1Id]
	assert.Len(s.T(), integration.Apps, 0)

	err = s.appDataGateway.DeleteById("dummy")
	assert.Error(s.T(), err, "application dummy not found")
}

func (s *testSuite) Test8_AG_FindWithRefresh() {
	integration := s.Data.Integrations[s.test1Id]
	assert.Len(s.T(), integration.Apps, 0)

	apps, err := s.appDataGateway.Find(true)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), apps, 1)

	assert.Len(s.T(), integration.Apps, 1)
}

func (s *testSuite) Test9_ConfigLoadExisting() {
	// this time load without filename
	testConfigPath := filepath.Join(s.testDir, ".hexa")

	_ = os.Setenv(EnvIntegrationConfigFile, testConfigPath)
	config, err := NewIntegrationConfigData()
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), config)
	assert.Len(s.T(), config.Integrations, 1)
}
