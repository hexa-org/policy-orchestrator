package testhelper

import "fmt"

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
