package microsoftazure

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"log"
	"net/http"
	"strings"
)

type AzureProvider struct {
	client AzureClient
}

type ProviderOpt func(provider *AzureProvider)

func WithAzureClient(clientOverride AzureClient) func(provider *AzureProvider) {
	return func(provider *AzureProvider) {
		provider.client = clientOverride
	}
}

func NewAzureProvider(opts ...ProviderOpt) *AzureProvider {
	provider := &AzureProvider{client: NewAzureClient(&http.Client{})}
	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

func (a *AzureProvider) Name() string {
	return "azure"
}

// DiscoverApplications For APIM
// ObjectID = service principal ID of Enterprise App
// Service = identifier (i.e. service URL) from App registration or APIM.serviceUrl
//
//	appReg.identifer == APIM.serviceUrl
//
// Name = APIM service name (also resource name)
//
//	apim.properties.serviceName == resource.name
//
// Description = APIM.properties.displayName
func (a *AzureProvider) DiscoverApplications(info orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	if !strings.EqualFold(info.Name, a.Name()) {
		return apps, err
	}

	key := info.Key
	found, _ := a.client.GetWebApplications(key)
	apps = append(apps, found...)
	return apps, err
}

func (a *AzureProvider) GetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	key := integrationInfo.Key
	servicePrincipals, _ := a.client.GetServicePrincipals(key, applicationInfo.Description) // todo - description is named poorly
	if len(servicePrincipals.List) == 0 {
		return []policysupport.PolicyInfo{}, nil
	}
	assignments, _ := a.client.GetAppRoleAssignedTo(key, servicePrincipals.List[0].ID)

	userEmailList := workflowsupport.ProcessAsync[AzureUser, AzureAppRoleAssignment](assignments.List, func(ara AzureAppRoleAssignment) (AzureUser, error) {
		user, _ := a.client.GetUserInfoFromPrincipalId(key, ara.PrincipalId)

		if user.Email == "" {
			return AzureUser{}, errors.New("no email found for principalId " + ara.PrincipalId)
		}
		return AzureUser{PrincipalId: ara.PrincipalId, Email: user.Email}, nil
	})

	userEmailMap := make(map[string]string)
	for _, ue := range userEmailList {
		if ue.PrincipalId != "" && ue.Email != "" {
			userEmailMap[ue.PrincipalId] = ue.Email
		}
	}

	policyMapper := NewAzurePolicyMapper(servicePrincipals, assignments.List, userEmailMap)
	return policyMapper.ToIDQL(), nil
}

func (a *AzureProvider) SetPolicyInfo(integrationInfo orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	validate := validator.New() // todo - move this up?
	errApp := validate.Struct(applicationInfo)
	if errApp != nil {
		return http.StatusInternalServerError, errApp
	}
	errPolicies := validate.Var(policyInfos, "omitempty,dive")
	if errPolicies != nil {
		return http.StatusInternalServerError, errPolicies
	}

	key := integrationInfo.Key

	sps, _ := a.client.GetServicePrincipals(key, applicationInfo.Description) // todo - description is named poorly

	appRoleValueToId := make(map[string]string)
	for _, ara := range sps.List[0].AppRoles {
		appRoleValueToId[ara.Value] = ara.ID
	}

	for _, policyInfo := range policyInfos {
		var assignments []AzureAppRoleAssignment

		actionUri := strings.TrimPrefix(policyInfo.Actions[0].ActionUri, "azure:")
		appRoleId, found := appRoleValueToId[actionUri]
		if !found {
			log.Println("No Azure AppRoleAssignment found for policy action", actionUri)
			continue
		}

		if len(policyInfo.Subject.Members) == 0 {
			assignments = append(assignments, AzureAppRoleAssignment{
				AppRoleId:  appRoleId,
				ResourceId: sps.List[0].ID,
				//ResourceId: strings.Split(policyInfo.Object.ResourceID, ":")[0],

			})
		}

		for _, user := range policyInfo.Subject.Members {
			principalId, _ := a.client.GetPrincipalIdFromEmail(key, strings.Split(user, ":")[1])
			if principalId == "" {
				continue
			}
			assignments = append(assignments, AzureAppRoleAssignment{
				AppRoleId:   appRoleId,
				PrincipalId: principalId,
				ResourceId:  sps.List[0].ID,
				//ResourceId:  strings.Split(policyInfo.Object.ResourceID, ":")[0],
			})
		}

		if len(assignments) == 0 {
			continue
		}

		err := a.client.SetAppRoleAssignedTo(key, sps.List[0].ID, assignments)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusCreated, nil
}
