package awsapigw

import (
	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw/dynamodbpolicy"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscognito"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
	log "golang.org/x/exp/slog"
	"net/http"
	"strings"
)

type AwsApiGatewayProvider struct {
	cognitoClientOverride  awscognito.CognitoClient
	policyStoreSvcOverride dynamodbpolicy.PolicyStoreSvc
	hasOverrides           bool
}

type AwsApiGatewayProviderOpt func(provider *AwsApiGatewayProvider)

func WithCognitoClientOverride(cognitoClientOverride awscognito.CognitoClient) AwsApiGatewayProviderOpt {
	return func(provider *AwsApiGatewayProvider) {
		provider.cognitoClientOverride = cognitoClientOverride
		provider.hasOverrides = true
	}
}

func WithPolicyStoreSvcOverride(policyStoreSvcOverride dynamodbpolicy.PolicyStoreSvc) AwsApiGatewayProviderOpt {
	return func(provider *AwsApiGatewayProvider) {
		provider.policyStoreSvcOverride = policyStoreSvcOverride
		provider.hasOverrides = true
	}
}
func NewAwsApiGatewayProvider(opts ...AwsApiGatewayProviderOpt) *AwsApiGatewayProvider {
	provider := &AwsApiGatewayProvider{}
	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

func (a *AwsApiGatewayProvider) Name() string {
	return "amazon"
}

func (a *AwsApiGatewayProvider) DiscoverApplications(integrationInfo orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	log.Info("AwsApiGatewayProvider.DiscoverApplications", "info.Name", integrationInfo.Name, "a.Name", a.Name())
	if !strings.EqualFold(integrationInfo.Name, a.Name()) {
		return []orchestrator.ApplicationInfo{}, err
	}
	service, err := a.getProviderService(integrationInfo.Key)
	if err != nil {
		log.Error("AwsApiGatewayProvider.DiscoverApplications", "getProviderService err", err)
		return []orchestrator.ApplicationInfo{}, err
	}

	return service.DiscoverApplications(integrationInfo)
}

func (a *AwsApiGatewayProvider) GetPolicyInfo(info orchestrator.IntegrationInfo, appInfo orchestrator.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	service, err := a.getProviderService(info.Key)
	if err != nil {
		log.Error("AwsApiGatewayProvider.GetPolicyInfo", "getProviderService err", err)
		return []hexapolicy.PolicyInfo{}, err
	}
	return service.GetPolicyInfo(appInfo)
}

func (a *AwsApiGatewayProvider) SetPolicyInfo(info orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (status int, foundErr error) {
	validate := validator.New()
	err := validate.Struct(applicationInfo)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = validate.Var(policyInfos, "omitempty,dive")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	service, err := a.getProviderService(info.Key)
	if err != nil {
		log.Error("AwsApiGatewayProvider.SetPolicyInfo", "getProviderService err", err)
		return http.StatusBadGateway, nil
	}

	return service.SetPolicyInfo(applicationInfo, policyInfos)
}

func (a *AwsApiGatewayProvider) getProviderService(key []byte) (*AwsApiGatewayProviderService, error) {
	var cognitoClient awscognito.CognitoClient
	var policyStoreSvc dynamodbpolicy.PolicyStoreSvc
	var err error

	if a.hasOverrides {
		cognitoClient = a.cognitoClientOverride
		policyStoreSvc = a.policyStoreSvcOverride
	} else {
		cognitoClient, err = awscognito.NewCognitoClient(key, awscommon.AWSClientOptions{})
		if err != nil {
			return nil, err
		}
		dynamodbClient, err := dynamodbpolicy.NewDynamodbClient(key, awscommon.AWSClientOptions{})
		if err != nil {
			return nil, err
		}

		policyStoreSvc = dynamodbpolicy.NewPolicyStoreSvc(dynamodbClient)
	}

	return NewAwsApiGatewayProviderService(cognitoClient, policyStoreSvc), nil
}
