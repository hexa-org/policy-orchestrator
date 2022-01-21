package provider

type Provider interface {
	Name() string
	DiscoverApplications(IntegrationInfo) ([]ApplicationInfo, error)
	GetPolicyInfo(IntegrationInfo, ApplicationInfo) ([]PolicyInfo, error)
	SetPolicyInfo(IntegrationInfo, ApplicationInfo, PolicyInfo) error
}

type IntegrationInfo struct {
	Name string
	Key  []byte
}

type ApplicationInfo struct {
	ObjectID    string
	Name        string
	Description string
}

type PolicyInfo struct {
	Version string
	Action  string
	Subject SubjectInfo
	Object  ObjectInfo
}

type SubjectInfo struct {
	AuthenticatedUsers []string
}

type ObjectInfo struct {
	Resources []string
}
