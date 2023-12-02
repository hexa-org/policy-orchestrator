package providercognito

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/providersV2/apps"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp"
)

type cognitoIdp struct {
	name string
	key  []byte
}

func NewCognitoIdp(key []byte) apps.Idp {
	return &cognitoIdp{key: key}
}

func (c *cognitoIdp) Provider() (idp.AppInfoSvc, error) {
	return cognitoidp.NewAppInfoSvc(c.key)
}
