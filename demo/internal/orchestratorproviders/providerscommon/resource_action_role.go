package providerscommon

import (
	"encoding/json"
	log "golang.org/x/exp/slog"
	"net/http"
	"strings"
)

const rarSeparator = "-"
const rarNVPrefix = "resrol"
const providerKeyActionPrefix = "http"

type ResourceActionRoles struct {
	Action   string // http method e.g GET
	Resource string
	Roles    []string
}

func NewResourceActionRoles(resource, httpMethod string, roles []string) ResourceActionRoles {
	if getHttpMethod(httpMethod, "") == "" {
		log.Warn("NewResourceActionRoles could not resolve httpMethod", "httpMethod", httpMethod, "resource", resource)
		return ResourceActionRoles{}
	}

	return newResourceActionRoles(resource, httpMethod, roles)
}

func NewResourceActionUriRoles(resource, actionUri string, roles []string) ResourceActionRoles {
	httpMethod := getHttpMethod(actionUri, ActionUriPrefix)
	if httpMethod == "" {
		log.Warn("NewResourceActionUriRoles could not resolve httpMethod", "action", actionUri, "resource", resource)
		return ResourceActionRoles{}
	}

	return newResourceActionRoles(resource, httpMethod, roles)
}

// NewResourceActionRolesFromProviderValue
// Build ResourceActionRoles from resAction, roles from provider
func NewResourceActionRolesFromProviderValue(resActionKey string, roles []string) ResourceActionRoles {
	parts := strings.Split(resActionKey, rarSeparator)
	if len(parts) < 3 {
		return ResourceActionRoles{}
	}
	prefix := parts[0]
	if prefix != rarNVPrefix {
		return ResourceActionRoles{}
	}

	action := parts[1]
	httpMethod := getHttpMethod(action, providerKeyActionPrefix)
	if httpMethod == "" {
		log.Warn("NewResourceActionRolesFromProviderValue could not resolve httpMethod", "resActionKey", resActionKey)
		return ResourceActionRoles{}
	}

	resource := strings.Join(parts[2:], "/")
	return newResourceActionRoles(resource, httpMethod, roles)
}

func newResourceActionRoles(resource, action string, roles []string) ResourceActionRoles {
	// TODO - Check if resource itself is "/"
	if strings.TrimSpace(resource) == "" || strings.TrimSpace(action) == "" {
		log.Warn("newResourceActionRoles empty resource or action", "resource", resource, "action", action)
		return ResourceActionRoles{}
	}

	res := strings.TrimSpace(resource)
	res, _ = strings.CutPrefix(res, "/")

	return ResourceActionRoles{
		Action:   action,
		Resource: "/" + strings.TrimSpace(res),
		Roles:    SanitizeMembers(roles),
	}
}

// Name
// see makeRarKey
func (nv ResourceActionRoles) Name() string {
	return makeRarKey(nv.Action, nv.Resource, "")
}

// Value
// returns a json string representing the roles array
func (nv ResourceActionRoles) Value() string {
	nvVal, _ := json.Marshal(nv.Roles)
	return string(nvVal)
}

// MakeRarKeyForPolicy
// convert policy actionUri to rarKey e.g. "resrol-httpget-humanresources-us"
func MakeRarKeyForPolicy(actionUri, resource string) string {
	return makeRarKey(actionUri, resource, ActionUriPrefix)
}

func SanitizeMembers(members []string) []string {
	sanitizedMembers := make([]string, 0)
	for _, mem := range members {
		if strings.TrimSpace(mem) == "" {
			continue
		}
		sanitizedMembers = append(sanitizedMembers, strings.TrimSpace(mem))
	}
	return sanitizedMembers
}

func makeRarKey(action, resource, actionPrefix string) string {
	// TODO - Check if resource itself is "/"
	if strings.TrimSpace(resource) == "" || strings.TrimSpace(action) == "" {
		log.Warn("makeRarKey empty resource or action", "resource", resource, "action", action)
		return ""
	}
	
	resNoPrefix, _ := strings.CutPrefix(resource, "/")
	httpMethod := getHttpMethod(action, actionPrefix)
	if httpMethod == "" {
		log.Warn("MakeRarKey could not resolve httpMethod", "action", action, "resource", resource)
		return ""
	}

	parts := []string{
		rarNVPrefix,
		providerKeyActionPrefix + strings.ToLower(httpMethod),
		strings.ReplaceAll(resNoPrefix, "/", rarSeparator),
	}

	nvName := strings.Join(parts, rarSeparator)
	return nvName
}

// getHttpMethod - converts an action ("httpget") or actionUri("http:GET") e.g.  to the
// corresponding http method i.e. GET
func getHttpMethod(action, actionPrefix string) string {
	for _, httpMethod := range []string{http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace} {

		prefixedHttpMethod := actionPrefix + httpMethod
		if strings.ToLower(prefixedHttpMethod) == strings.ToLower(action) {
			return httpMethod
		}
	}
	return ""
}
