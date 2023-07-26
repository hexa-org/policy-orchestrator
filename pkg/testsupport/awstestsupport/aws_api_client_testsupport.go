package awstestsupport

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
)

const TestAwsRegion = "us-west-1"
const TestAwsAccessKeyId = "anAccessKeyID"
const TestAwsSecretAccessKey = "aSecretAccessKey"

const TestUserPoolId = "some-user-pool-id"
const TestUserPoolName = "some-user-pool-name"
const TestResourceServerIdentifier = "https://some-resource-server"
const TestResourceServerName = "some-resource-server-name"

func AwsCredentialsForTest() []byte {
	str := fmt.Sprintf(`
{
  "accessKeyID": "%s",
  "secretAccessKey": "%s",
  "region": "%s"
}
`, TestAwsAccessKeyId, TestAwsSecretAccessKey, TestAwsRegion)

	return []byte(str)
}

func IntegrationInfo() orchestrator.IntegrationInfo {
	return orchestrator.IntegrationInfo{Name: "amazon", Key: AwsCredentialsForTest()}
}

func AppInfo() orchestrator.ApplicationInfo {
	return orchestrator.ApplicationInfo{
		ObjectID:    TestUserPoolId,
		Name:        TestResourceServerName,
		Description: "Cognito",
		Service:     TestResourceServerIdentifier,
	}
}
