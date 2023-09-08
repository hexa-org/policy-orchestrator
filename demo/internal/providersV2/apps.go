package providersV2

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
)

type Idp interface {
	Provider() (idp.AppInfoSvc, error)
}
