module github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore

go 1.20

// Temporary replace till we commit
// SAURABH replace github.com/hexa-org/policy-orchestrator/sdk/core => ./../../../core

// SAURABH replace github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon => ../../../provideraws/awscommon

require (
	github.com/aws/aws-sdk-go-v2 v1.21.2
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.10.42
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.22.2
	github.com/aws/smithy-go v1.15.0
	// SAURABH github.com/hexa-org/policy-orchestrator/sdk/core v0.0.0-00010101000000-000000000000
	// SAURABH github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.4
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
)

require (
	github.com/aws/aws-sdk-go-v2/config v1.18.45 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.43 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.45 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.15.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.23.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hexa-org/policy-mapper/hexaIdql v0.6.0-alpha.3 // indirect
	github.com/hexa-org/policy-orchestrator/sdk/core v0.6.0-alpha.1 // indirect
	github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon v0.6.0-alpha.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)