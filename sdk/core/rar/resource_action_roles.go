package rar

import (
	"errors"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"golang.org/x/exp/slices"
	log "golang.org/x/exp/slog"
	"net/http"
	"strings"
)

const ActionUriPrefix = "http:"

var supportedHttpMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost,
	http.MethodPut, http.MethodPatch, http.MethodDelete,
	http.MethodConnect, http.MethodOptions, http.MethodTrace}

// ResourceActionRolesMapper - Clients provide implementation based on their policy schema
// The external vendor specific policy struct must implement this interface
// which will be used to convert the vendor specific policy to ResourceActionRoles
// Simple mapper with
// - non composite keys and values
// auto generated table definition
type ResourceActionRolesMapper interface {
	MapTo() (ResourceActionRoles, error)
}

type DynamicResourceActionRolesMapper struct {
}

func (d DynamicResourceActionRolesMapper) MapTo() (ResourceActionRoles, error) {
	// TODO implement me
	panic("implement me")
}

// ResourceActionRoles - an internal representation of a policy
// Vendor specific policies are transformed to / from IDQL
// using this struct
// TODO - Rename to something better
type ResourceActionRoles struct {
	resource string
	actions  []string // http method e.g GET
	roles    []string
}

// NewResourceActionRoles - creates ResourceActionRoles with specified
// resource, http methods and roles
// e.g. NewResourceActionRoles("some-resource", {http.GET, http.POST}, {...} )
func NewResourceActionRoles(resource string, httpMethods []string, roles []string) (ResourceActionRoles, error) {
	return newResourceActionRoles(resource, httpMethods, roles)
}

// NewResourceActionUriRoles creates ResourceActionRoles for an idql policy
// e.g. NewResourceActionUriRoles("some-resource", {"http:GET", "http:POST"}, {...} )
func NewResourceActionUriRoles(resource string, actionUris []string, roles []string) (ResourceActionRoles, error) {
	httpMethods := make([]string, 0)
	for _, prefixedMethod := range actionUris {
		trimmed := strings.TrimSpace(prefixedMethod)
		unPrefixed := strings.TrimPrefix(trimmed, ActionUriPrefix)
		httpMethods = append(httpMethods, unPrefixed)
	}
	return newResourceActionRoles(resource, httpMethods, roles)
}

func newResourceActionRoles(aResource string, httpMethods []string, roles []string) (ResourceActionRoles, error) {

	resource := strings.TrimSpace(aResource)
	// TODO - Check if resource itself is "/"
	if resource == "" {
		log.Warn("NewResourceActionRoles empty resource")
		return ResourceActionRoles{}, errors.New("error creating ResourceActionRole with empty resource")
	}

	tmpActions := make([]string, 0)
	for _, aMethod := range httpMethods {
		trimmed := strings.TrimSpace(aMethod)
		if slices.Index(supportedHttpMethods, trimmed) < 0 {
			return ResourceActionRoles{}, errors.New("error creating ResourceActionRole. Invalid http method " + aMethod)
		}
		tmpActions = append(tmpActions, trimmed)
	}

	sortedActions := sanitizeAndSort(tmpActions)
	members := sanitizeAndSort(roles)
	return ResourceActionRoles{resource: resource,
		actions: sortedActions, roles: members}, nil
}

func (rar ResourceActionRoles) Resource() string {
	return rar.resource
}

func (rar ResourceActionRoles) Actions() []string {
	return rar.actions
}

func (rar ResourceActionRoles) Members() []string {
	return rar.roles
}

func (rar ResourceActionRoles) ToIDQL() hexapolicy.PolicyInfo {
	actionInfos := make([]hexapolicy.ActionInfo, 0)
	for _, act := range rar.Actions() {
		actionInfos = append(actionInfos, hexapolicy.ActionInfo{ActionUri: ActionUriPrefix + act})
	}
	return hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: actionInfos,
		Subject: hexapolicy.SubjectInfo{Members: rar.Members()},
		Object:  hexapolicy.ObjectInfo{ResourceID: rar.Resource()},
	}
}

// sanitizeAndSort - removes duplicates, trims each element
// returns sorted slice
func sanitizeAndSort(orig []string) []string {
	slices.Sort(orig) // keep sorted, and also compact replaces consecutive elements
	ret := make([]string, 0)
	// Compact to remove duplicates
	for _, elem := range slices.Compact(orig) {
		aElem := strings.TrimSpace(elem)
		if len(aElem) > 0 {
			ret = append(ret, aElem)
		}
	}

	return ret
}
