package policyprovider

import (
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policystore"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	log "golang.org/x/exp/slog"
)

type ProviderService interface {
	DiscoverApplications() ([]idp.AppInfo, error)
	GetPolicyInfo(idp.AppInfo) ([]hexapolicy.PolicyInfo, error)
	SetPolicyInfo(idp.AppInfo, []hexapolicy.PolicyInfo) error
}

type providerService[R any] struct {
	appInfoSvc     idp.AppInfoSvc
	policyStoreSvc policystore.PolicyBackendSvc[R]
}

func NewProviderService[R any](appInfoService idp.AppInfoSvc, policyStoreSvc policystore.PolicyBackendSvc[R]) ProviderService {
	return &providerService[R]{appInfoSvc: appInfoService, policyStoreSvc: policyStoreSvc}
}

func (s *providerService[R]) DiscoverApplications() ([]idp.AppInfo, error) {
	return s.appInfoSvc.GetApplications()
}
func (s *providerService[R]) GetPolicyInfo(appInfo idp.AppInfo) ([]hexapolicy.PolicyInfo, error) {
	rarList, err := s.policyStoreSvc.GetPolicies(appInfo)
	if err != nil {
		log.Error("ProviderService.GetPolicyInfo",
			"failed calling GetPolicies App.Name", appInfo.Name(),
			"App.Id", appInfo.Id(),
			"err=", err)
		return []hexapolicy.PolicyInfo{}, err
	}
	return buildPolicies(rarList), nil
}
func (s *providerService[R]) SetPolicyInfo(appInfo idp.AppInfo, policies []hexapolicy.PolicyInfo) error {
	log.Info("policyprovider.ProviderService", "appInfo", appInfo)
	log.Info("policyprovider.ProviderService", "policies", policies)
	existingRarList, err := s.policyStoreSvc.GetPolicies(appInfo)
	if err != nil {
		log.Error("ProviderService.SetPolicyInfo",
			"failed calling GetPolicies App.Name", appInfo.Name(),
			"App.Id", appInfo.Id(),
			"err=", err)
		return err
	}

	log.Info("policyprovider.ProviderService", "existingRarList", existingRarList)

	if len(existingRarList) == 0 {
		log.Info("ProviderService.SetPolicyInfo", "no existing policies, returning", "appInfo.Name()", appInfo.Name())
		return nil
	}

	newPoliciesRarMap, err := mapIdqlToRar(policies...)
	log.Info("policyprovider.ProviderService", "newPoliciesRarMap", newPoliciesRarMap)
	if err != nil {
		log.Error("ProviderService.SetPolicyInfo",
			"failed to map IDQL to rar", appInfo.Name(),
			"err=", err)
		return err
	}

	if len(newPoliciesRarMap) == 0 {
		log.Info("ProviderService.SetPolicyInfo", "no new policies, returning", "appInfo.Name()", appInfo.Name())
		return nil
	}

	updateCalc := newUpdateCalculator(existingRarList, newPoliciesRarMap)
	updateList := updateCalc.calculate()

	for _, aRar := range updateList {
		log.Info("ProviderService.SetPolicyInfo", "msg", "call policyStoreSvc.SetPolicy", "aRar", aRar)
		updateErr := s.policyStoreSvc.SetPolicy(aRar)
		if updateErr != nil {
			log.Error("ProviderService.SetPolicyInfo", "msg", "failed to update policy in backend store",
				"resource", aRar.Resource(),
				"action", aRar.Actions(),
				"members", aRar.Members(),
				"error", updateErr)
			return updateErr
		}
	}

	return nil // StatusCreated
}

func Hello(name string) string {
	return "Hello " + name
}

func buildPolicies(rarList []rar.ResourceActionRoles) []hexapolicy.PolicyInfo {
	policies := make([]hexapolicy.PolicyInfo, 0)
	for _, aRar := range rarList {
		policies = append(policies, aRar.ToIDQL())
	}
	return policies
}
