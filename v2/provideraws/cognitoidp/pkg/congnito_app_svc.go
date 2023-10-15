package cognitoidp

import (
	"github.com/hexa-org/policy-orchestrator/v2/core/idp"
	logger "golang.org/x/exp/slog"
)

type appInfoSvc struct {
	client CognitoClient
}

type Opt func(svc *appInfoSvc)

func WithCognitoClientOverride(client CognitoClient) Opt {
	return func(svc *appInfoSvc) {
		svc.client = client
	}
}
func NewAppInfoSvc(key []byte, opts ...Opt) (idp.AppInfoSvc, error) {
	if len(opts) == 0 {
		client, err := NewCognitoClient(key, nil)
		if err != nil {
			logger.Error("NewAppInfoSvc", "error building CognitoClient", "error", err.Error())
			return nil, err
		}
		return &appInfoSvc{client: client}, nil
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
	pools, err := as.client.listUserPools()
	if err != nil {
		logger.Error("getResourceServers", "error calling listUserPools aws cognito api", "error", err.Error())
		return nil, err
	}

	apps := make([]idp.AppInfo, 0)
	for _, p := range pools.UserPools {
		rsOutput, err := as.client.listResourceServers(*p.Id)
		if err != nil {
			logger.Error("getResourceServers", "error calling listResourceServers aws cognito api", "UserPoolId", *p.Id, "error", err.Error())
			return nil, err
		}

		for _, rs := range rsOutput.ResourceServers {
			apps = append(apps, NewResourceServerAppInfo(*rs.UserPoolId, *rs.Name, *rs.Name, *rs.Identifier))
		}
	}
	return apps, err
}
