module github.com/hexa-org/policy-orchestrator/v2/provideraws/app/cognitoidp

go 1.20

// temporary replace till we push to github and tag
replace github.com/hexa-org/policy-orchestrator/v2/core => ./../../core

require (
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.27.0
	github.com/hexa-org/policy-orchestrator/v2/core v0.0.0-00010101000000-000000000000
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9
)

require (
	github.com/aws/aws-sdk-go-v2 v1.21.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/smithy-go v1.15.0 // indirect
)
