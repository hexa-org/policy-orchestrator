package apigwprovider

import (
	"errors"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/v2/core/idp"
	"github.com/hexa-org/policy-orchestrator/v2/core/policyprovider"
	"github.com/hexa-org/policy-orchestrator/v2/core/policystore"
	log "golang.org/x/exp/slog"
)

type ProviderService struct {
	appInfoSvc     idp.AppInfoSvc
	policyStoreSvc policystore.PolicyStoreSvc
}

func NewProviderService(appInfoService idp.AppInfoSvc, policyStoreSvc policystore.PolicyStoreSvc) policyprovider.ProviderService {
	return &ProviderService{appInfoSvc: appInfoService, policyStoreSvc: policyStoreSvc}
}

func (s *ProviderService) DiscoverApplications() ([]idp.AppInfo, error) {
	return s.appInfoSvc.GetApplications()
}
func (s *ProviderService) GetPolicyInfo(appInfo idp.AppInfo) ([]hexapolicy.PolicyInfo, error) {
	log.Debug("ProviderService", "begin GetPolicyInfo", "appInfo.Id", appInfo.Id(), "appInfo.Name", appInfo.Name())
	resServerAppInfo, ok := appInfo.(pkg.ResourceServerAppInfo)
	if !ok {
		log.Error("ProviderService", "GetPolicyInfo", "invalid appInfo", "expecting ResourceServerAppInfo")
		return nil, errors.New("ProviderService.GetPolicyInfo invalid appInfo type")
	}
	_, err := s.policyStoreSvc.GetPolicies(resServerAppInfo)
	if err != nil {
		log.Error("ProviderService.GetPolicyInfo", "error calling GetResourceRoles App.Name", appInfo.Name, "identifierUrl", resServerAppInfo.Identifier(), "err=", err)
		return []hexapolicy.PolicyInfo{}, err
	}
	return nil, nil
}
func (s *ProviderService) SetPolicyInfo(appInfo idp.AppInfo, policies []hexapolicy.PolicyInfo) error {
	// convert policies to rars and call SetPolicy
	// s.policyStoreSvc.SetPolicy(policies)
	return nil // StatusCreated
}
