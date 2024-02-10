package idp

type AppInfo interface {
	Id() string
	Name() string
	DisplayName() string
	Type() string
}

type AppInfoSvc interface {
	GetApplications() ([]AppInfo, error)
	// GetApplication - key can mean different things to different providers
	GetApplication(key string) (AppInfo, error)
}
