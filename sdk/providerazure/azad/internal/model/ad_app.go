package model

// ResourceServerAppInfo - consider moving this to core
// same struct used by cognito
type ResourceServerAppInfo struct {
	id          string
	name        string
	description string
	identifier  string
}

func NewResourceServerAppInfo(id string, name string, description string, identifier string) ResourceServerAppInfo {
	return ResourceServerAppInfo{id: id, name: name, description: description, identifier: identifier}
}

func (a ResourceServerAppInfo) Id() string {
	return a.id
}
func (a ResourceServerAppInfo) Name() string {
	return a.name
}
func (a ResourceServerAppInfo) DisplayName() string {
	return a.description
}
func (a ResourceServerAppInfo) Identifier() string {
	return a.identifier
}
func (a ResourceServerAppInfo) Type() string {
	return "Azure AD Application"
}
