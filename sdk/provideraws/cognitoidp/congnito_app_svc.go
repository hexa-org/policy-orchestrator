package cognitoidp

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp/internal/client"
	logger "golang.org/x/exp/slog"
)

type appInfoSvc struct {
	cognito client.CognitoClient
}

type Opt func(svc *appInfoSvc)

func WithCognitoClientOverride(client client.CognitoClient) Opt {
	return func(svc *appInfoSvc) {
		svc.cognito = client
	}
}
func NewAppInfoSvc(key []byte, opts ...Opt) (idp.AppInfoSvc, error) {
	if len(opts) == 0 {
		cognito, err := client.NewCognitoClient(key, nil)
		if err != nil {
			logger.Error("NewAppInfoSvc", "error building CognitoClient", "error", err.Error())
			return nil, err
		}
		return &appInfoSvc{cognito: cognito}, nil
	}

	svc := &appInfoSvc{}
	for _, o := range opts {
		o(svc)
	}
	return svc, nil
}

func (as *appInfoSvc) GetApplications() ([]idp.AppInfo, error) {
	return as.getResourceServers()
}

func (as *appInfoSvc) getResourceServers() ([]idp.AppInfo, error) {
	pools, err := as.cognito.ListUserPools()
	if err != nil {
		logger.Error("getResourceServers", "error calling listUserPools aws cognito api", err.Error())
		return nil, err
	}

	apps := make([]idp.AppInfo, 0)
	for _, p := range pools.UserPools {
		rsOutput, err := as.cognito.ListResourceServers(*p.Id)
		if err != nil {
			logger.Error("getResourceServers", "error calling listResourceServers aws cognito api. UserPoolId", *p.Id, "error", err.Error())
			return nil, err
		}

		for _, rs := range rsOutput.ResourceServers {
			apps = append(apps, NewResourceServerAppInfo(*rs.UserPoolId, *rs.Name, *rs.Name, *rs.Identifier))
		}
	}
	return apps, err
}
