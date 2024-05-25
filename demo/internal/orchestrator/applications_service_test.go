package orchestrator_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	assert "github.com/stretchr/testify/require"
)

type applicationsServiceData struct {
	// db        *sql.DB
	providers map[string]policyprovider.Provider

	fromNoopApp string
	toNoopApp   string
	azureApp    string
	googleApp   string
	testDir     string
	Data        *dataConfigGateway.ConfigData
	appGateway  dataConfigGateway.ApplicationsDataGateway
	intGateway  dataConfigGateway.IntegrationsDataGateway
}

func (data *applicationsServiceData) SetUp() {
	tempDir, _ := os.MkdirTemp("", "hexa-orchestrator-*")

	data.testDir = tempDir

	testConfigPath := filepath.Join(data.testDir, ".hexa", "config.json")
	_ = os.Setenv(sdk.EnvTestProvider, "noop")
	// _ = os.Unsetenv(sdk.EnvTestProvider)

	_ = os.Setenv(sdk.EnvTestProvider, sdk.ProviderTypeMock)
	_ = os.Setenv(dataConfigGateway.EnvIntegrationConfigFile, testConfigPath)

	data.Data, _ = dataConfigGateway.NewIntegrationConfigData()
	data.appGateway = data.Data.GetApplicationDataGateway()
	data.intGateway = data.Data

	/*
	   	data.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	   	_, err := data.db.Exec(`
	   delete from applications;
	   delete from integrations;
	   insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'aName', 'noop', 'aKey');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff210', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anObjectId', 'aName', 'aDescription', 'aService');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff211', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anotherObjectId', 'anotherName', 'anotherDescription', 'anotherService');

	   insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e8', 'anotherName', 'azure_legacy', 'aKey');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff212', '50e00619-9f15-4e85-a7e9-f26d87ea12e8', 'andAnotherObjectId', 'andAnotherName', 'andAnotherDescription', 'andAnotherService');

	   insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e9', 'yetAnotherName', 'google_cloud', 'aKey');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff213', '50e00619-9f15-4e85-a7e9-f26d87ea12e9', 'yetAnotherObjectId', 'yetAnotherName', 'yetAnotherDescription', 'yetAnotherService');
	   `)



	   	// TODO: pass `testing.T` into this `SetUp`.
	   	// e.g., so we can `assert.NoError(t, ...)`, `t.Fatal()` instead of...
	   	if err != nil {
	   		log.Fatalf("failed to setup ConfigData: %v", err)
	   	}

	*/
	_, _ = data.Data.Create("50e00619-9f15-4e85-a7e9-f26d87ea12e7", "noop", []byte("aKey"))
	integration := data.Data.Integrations["50e00619-9f15-4e85-a7e9-f26d87ea12e7"]
	apps := []policyprovider.ApplicationInfo{
		{"anObjectId", "aName", "aDescription", "aService"},
		{"anotherObjectId", "anotherName", "anotherDescription", "anotherService"},
	}
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff210"] = apps[0]
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff211"] = apps[0]

	_, _ = data.Data.Create("50e00619-9f15-4e85-a7e9-f26d87ea12e8", "noop", []byte("aKey"))
	integration = data.Data.Integrations["50e00619-9f15-4e85-a7e9-f26d87ea12e8"]
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff212"] = policyprovider.ApplicationInfo{"andAnotherObjectId", "andAnotherName", "andAnotherDescription", "andAnotherService"}
	_, _ = data.Data.Create("50e00619-9f15-4e85-a7e9-f26d87ea12e9", "noop", []byte("aKey"))
	integration = data.Data.Integrations["50e00619-9f15-4e85-a7e9-f26d87ea12e8"]
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff213"] = policyprovider.ApplicationInfo{"yetAnotherObjectId", "yetAnotherName", "yetAnotherDescription", "yetAnotherService"}

	data.fromNoopApp = "6409776a-367a-483a-a194-5ccf9c4ff210"
	data.toNoopApp = "6409776a-367a-483a-a194-5ccf9c4ff210"
	data.azureApp = "6409776a-367a-483a-a194-5ccf9c4ff212"
	data.googleApp = "6409776a-367a-483a-a194-5ccf9c4ff213"

	data.providers = make(map[string]policyprovider.Provider)
	data.providers["50e00619-9f15-4e85-a7e9-f26d87ea12e7"] = &orchestrator_test.NoopProvider{}
	data.providers["50e00619-9f15-4e85-a7e9-f26d87ea12e8"] = &orchestrator_test.NoopProvider{OverrideName: "azure_legacy"}
	data.providers["50e00619-9f15-4e85-a7e9-f26d87ea12e9"] = &orchestrator_test.NoopProvider{OverrideName: "google_cloud"}
}

func (data *applicationsServiceData) TearDown() {
	// _ = data.db.Close()
	_ = os.RemoveAll(data.testDir)
}

func TestApplicationsService_Apply(t *testing.T) {
	testsupport.WithSetUp(&applicationsServiceData{}, func(data *applicationsServiceData) {
		pb := orchestrator.NewProviderBuilder()
		pb.AddProviders(data.providers)
		applicationsGateway := data.appGateway
		integrationsGateway := data.intGateway

		applicationsService := orchestrator.ApplicationsService{ApplicationsGateway: applicationsGateway, IntegrationsGateway: integrationsGateway, ProviderBuilder: pb, DisableChecks: true}

		sameProviders := applicationsService.Apply(orchestrator.Orchestration{From: data.fromNoopApp, To: data.toNoopApp})
		assert.NoError(t, sameProviders)

		differentProviders := applicationsService.Apply(orchestrator.Orchestration{From: data.azureApp, To: data.googleApp})
		assert.NoError(t, differentProviders)

		badFromApp := applicationsService.Apply(orchestrator.Orchestration{From: "", To: data.toNoopApp})
		assert.Error(t, badFromApp)

		badToApp := applicationsService.Apply(orchestrator.Orchestration{From: data.fromNoopApp, To: ""})
		assert.Error(t, badToApp)

		noopProvider := data.providers["50e00619-9f15-4e85-a7e9-f26d87ea12e7"].(*orchestrator_test.NoopProvider)
		noopProvider.SetTestErr(errors.New("oops"))

		providerError := applicationsService.Apply(orchestrator.Orchestration{From: data.fromNoopApp, To: data.toNoopApp})
		assert.Error(t, providerError)

		// reset after test
		noopProvider.SetTestErr(nil)
	})
}

func TestApplicationsService_RetainResource(t *testing.T) {
	testsupport.WithSetUp(&applicationsServiceData{}, func(data *applicationsServiceData) {

		applicationsService := orchestrator.ApplicationsService{ApplicationsGateway: data.appGateway, IntegrationsGateway: data.intGateway, ProviderBuilder: nil}

		from := []hexapolicy.PolicyInfo{
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"fromAnAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"fromAUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "fromAnId",
			}},
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"fromAnotherAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"fromAnotherUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "fromAnId",
			}},
		}

		to := []hexapolicy.PolicyInfo{
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"toAnAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"toAUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "toAnId",
			}},
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"toAnotherAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"toAnotherUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "toAnId",
			}},
		}

		modified, _ := applicationsService.RetainResource(from, to)
		assert.Equal(t, "toAnId", modified[0].Object.ResourceID)
		assert.Equal(t, "fromAUser", modified[0].Subject.Members[0])

		assert.Equal(t, "toAnId", modified[1].Object.ResourceID)
		assert.Equal(t, "fromAnotherUser", modified[1].Subject.Members[0])

		toWithDifferentResources := []hexapolicy.PolicyInfo{
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"anotherAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"anotherUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "anotherId",
			}},
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"anotherAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"anotherUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "andAnotherId",
			}},
		}
		_, err := applicationsService.RetainResource(from, toWithDifferentResources)
		assert.Error(t, err)
	})
}

func TestApplicationsService_RetainAction(t *testing.T) {
	testsupport.WithSetUp(&applicationsServiceData{}, func(data *applicationsServiceData) {

		applicationsService := orchestrator.ApplicationsService{ApplicationsGateway: data.appGateway, IntegrationsGateway: data.intGateway, ProviderBuilder: nil}

		from := []hexapolicy.PolicyInfo{
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"fromAnAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"fromAUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "fromAnId",
			}},
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"fromAnotherAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"fromAnotherUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "fromAnId",
			}},
		}

		to := []hexapolicy.PolicyInfo{
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"toAnAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"toAUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "toAnId",
			}},
			{Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"toAnotherAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"toAnotherUser"}}, Object: hexapolicy.ObjectInfo{
				ResourceID: "toAnId",
			}},
		}

		modified, _ := applicationsService.RetainAction(from, to)

		assert.Equal(t, "toAnAction", modified[0].Actions[0].ActionUri)

		assert.Equal(t, "toAnAction", modified[1].Actions[0].ActionUri)
	})
}
