package provider

type Provider interface {
	Name() string
	DiscoverApplications(IntegrationInfo) []ApplicationInfo
}

type IntegrationInfo struct {
	Name string
	Key []byte
}

type ApplicationInfo struct {
	Name string
}

type PolicyInfo struct {
	Name string
}
