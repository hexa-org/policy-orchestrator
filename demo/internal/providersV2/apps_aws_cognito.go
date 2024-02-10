package providersV2

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/cognitoidp"
)

type cognitoIdp struct {
	name string
	key  []byte
}

func NewCognitoIdp(key []byte) Idp {
	return &cognitoIdp{key: key}
}

func (c *cognitoIdp) Provider() (idp.AppInfoSvc, error) {
	return cognitoidp.NewAppInfoSvc(c.key)
}
