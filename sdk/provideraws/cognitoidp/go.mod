module github.com/hexa-org/policy-orchestrator/sdk/provideraws/app/cognitoidp

go 1.20

// temporary replace till we push to github and tag
//SAURABH TEMP replace github.com/hexa-org/policy-orchestrator/sdk/core => ./../../core

require (
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.27.0
	// SAURABH TEMP github.com/hexa-org/policy-orchestrator/sdk/core v0.0.0-00010101000000-000000000000
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
)

require (
	github.com/aws/aws-sdk-go-v2 v1.21.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/smithy-go v1.15.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hexa-org/policy-mapper/hexaIdql v0.6.0-alpha.3 // indirect
	github.com/hexa-org/policy-orchestrator/sdk/core v0.6.0-alpha.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
