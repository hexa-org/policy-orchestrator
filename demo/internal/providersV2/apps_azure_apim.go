package providersV2

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/apps"
)

type apimAppProvider struct {
	key []byte
}

func (a *apimAppProvider) Provider() (idp.AppInfoSvc, error) {
	return apps.NewAppInfoSvc(a.key, nil)
}

func NewApimAppProvider(key []byte) Idp {
	return &apimAppProvider{key: key}
}
