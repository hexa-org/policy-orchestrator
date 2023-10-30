module github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp

go 1.20

// temporary replace till we push to github and tag
//SAURABH TEMP replace github.com/hexa-org/policy-orchestrator/sdk/core => ./../../core

require (
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.27.0
	// SAURABH TEMP github.com/hexa-org/policy-orchestrator/sdk/core v0.0.0-00010101000000-000000000000
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
)

require (
	github.com/aws/aws-sdk-go-v2 v1.21.2
	github.com/hexa-org/policy-orchestrator/sdk/core v0.6.0-alpha.5
	github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon v0.6.0-alpha.5
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/aws/aws-sdk-go-v2/config v1.18.45 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.43 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.45 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.23.2 // indirect
	github.com/aws/smithy-go v1.15.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hexa-org/policy-mapper/hexaIdql v0.6.0-alpha.3 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)