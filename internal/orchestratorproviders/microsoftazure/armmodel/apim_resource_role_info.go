package armmodel

import (
	"net/http"
	"strings"
)

type ResourceActionRoles struct {
	ApimServiceInfo
	Action   string
	Resource string
	Roles    []string
}

func NewResourceActionRoles(resAction string, roles []string) ResourceActionRoles {
	parts := strings.Split(resAction, "-")
	if len(parts) < 3 {
		return ResourceActionRoles{}
	}
	prefix := parts[0]
	if prefix != "resrol" {
		return ResourceActionRoles{}
	}

	action := parts[1]
	httpMethod := getHttpMethod(action)
	if httpMethod == "" {
		return ResourceActionRoles{}
	}

	resource := strings.Join(parts[2:], "/")

	/*roles := make([]string, 0)
	for _, rol := range strings.Split(rolesStr, ",") {
		roles = append(roles, strings.TrimSpace(rol))
	}*/
	return ResourceActionRoles{
		Action:   httpMethod,
		Resource: "/" + resource,
		Roles:    roles,
	}
}

// getHttpMethod - converts an action e.g. "httpget" to the
//
//	corresponding http method i.e. GET
func getHttpMethod(action string) string {
	for _, httpMethod := range []string{http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace} {

		prefixedMethod := "http" + strings.ToLower(httpMethod)
		if prefixedMethod == action {
			return httpMethod
		}
	}
	return ""
}
