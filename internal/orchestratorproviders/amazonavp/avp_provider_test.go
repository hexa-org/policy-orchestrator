package amazonavp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonavp"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
	"github.com/stretchr/testify/assert"
)

type TestInfo struct {
	Apps     []orchestrator.ApplicationInfo
	Provider amazonavp.AmazonAvpProvider
	Info     orchestrator.IntegrationInfo
}

var testData TestInfo
var initialized = false

func initializeTests() error {
	if initialized {
		return nil
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	cred, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return err
	}

	str := fmt.Sprintf(`
{
  "accessKeyID": "%s",
  "secretAccessKey": "%s",
  "region": "%s"
}
`, cred.AccessKeyID, cred.SecretAccessKey, cfg.Region)

	info := orchestrator.IntegrationInfo{Name: "avp", Key: []byte(str)}
	avp := amazonavp.AmazonAvpProvider{AwsClientOpts: awscommon.AWSClientOptions{DisableRetry: true}}

	testData = TestInfo{
		Provider: avp,
		Info:     info,
	}

	initialized = true
	return nil
}

func TestAvp_1_DiscoverApplications(t *testing.T) {

	err := initializeTests()
	assert.NoError(t, err, "Should be initialized")

	apps, err := testData.Provider.DiscoverApplications(testData.Info)

	assert.NoError(t, err, "check no error")
	assert.NotNil(t, apps, "Apps not nil")

	fmt.Println("Apps:")
	fmt.Println(fmt.Sprintf("%s", apps))

	testData.Apps = apps
}

func TestAvp_2_ListPolicies(t *testing.T) {
	assert.NotNil(t, testData.Apps, "Apps should be initialized")

	for _, app := range testData.Apps {
		policies, err := testData.Provider.GetPolicyInfo(testData.Info, app)
		assert.NoError(t, err, "Get policy has no error")
		for _, hexaPol := range policies {
			polBytes, err := json.MarshalIndent(hexaPol, "", "  ")
			assert.NoError(t, err, "Policy should marshall")
			fmt.Println(fmt.Sprintf("Policy: \n%s", string(polBytes)))
		}
	}

}
