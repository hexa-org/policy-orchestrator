module github.com/hexa-org/policy-orchestrator/demo

go 1.20

// +heroku goVersion go1.20

// SAURABH TEMP REPACE FOR LOCAL DEV
//replace github.com/hexa-org/policy-orchestrator/sdk/core => ../sdk/core
//replace github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp => ../sdk/provideraws/cognitoidp
//replace github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore => ../sdk/provideraws/policystore/dynamodbpolicystore

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.4.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.2.2
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.1.1
	github.com/aws/aws-sdk-go-v2 v1.21.2
	github.com/aws/aws-sdk-go-v2/config v1.18.45
	github.com/aws/aws-sdk-go-v2/credentials v1.13.43
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.10.42
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.13.14
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.27.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.22.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.30.0
	github.com/envoyproxy/go-control-plane v0.11.1 // indirect
	github.com/go-playground/validator/v10 v10.11.2
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/cel-go v0.18.0 // indirect
	github.com/google/uuid v1.3.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	github.com/hiyosi/hawk v1.0.1
	github.com/lib/pq v1.10.7
	github.com/stretchr/testify v1.8.4
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	google.golang.org/api v0.139.0
	gopkg.in/square/go-jose.v2 v2.6.0
)

require (
	github.com/hexa-org/policy-mapper/hexaIdql v0.6.0-alpha.3
	github.com/hexa-org/policy-mapper/mapper/formats/gcpBind v0.6.0-alpha.3
	github.com/hexa-org/policy-orchestrator/sdk/core v0.6.0-alpha.8
	github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp v0.6.0-alpha.8
	github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore v0.6.0-alpha.8
)

require (
	github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon v0.6.0-alpha.8 // indirect
	golang.org/x/sync v0.4.0 // indirect
)

require (
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.2.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.9.0 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.45 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.15.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.23.2 // indirect
	github.com/aws/smithy-go v1.15.0 // indirect
	github.com/cncf/xds/go v0.0.0-20230607035331-e9ce68804cb4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.5 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/hexa-org/policy-mapper/mapper/conditionLangs/gcpcel v0.6.0-alpha.3 // indirect
	//github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon v0.6.0-alpha.8 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.16.0 // indirect
	golang.org/x/oauth2 v0.12.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.58.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
