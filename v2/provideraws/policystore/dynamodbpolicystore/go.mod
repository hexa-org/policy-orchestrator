module github.com/hexa-org/policy-orchestrator/v2/provideraws/policystore/dynamodbpolicystore

go 1.20

// Temporary replace till we commit
replace github.com/hexa-org/policy-orchestrator/v2/core => ./../../../core

replace github.com/hexa-org/policy-orchestrator/v2/provideraws/awscommon => ../../../provideraws/awscommon

require (
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.10.42
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.22.2
	github.com/hexa-org/policy-orchestrator/v2/core v0.0.0-00010101000000-000000000000
	github.com/hexa-org/policy-orchestrator/v2/provideraws/awscommon v0.0.0-00010101000000-000000000000
)

require (
	github.com/aws/aws-sdk-go-v2 v1.21.2 // indirect
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
	github.com/aws/smithy-go v1.15.0 // indirect
	github.com/hexa-org/policy-mapper/hexaIdql v0.6.0-alpha.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
)
