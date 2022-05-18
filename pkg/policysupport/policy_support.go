package policysupport

// todo - longer name used here to simplify a refactoring

type PolicyInfo struct {
	Version string
	Actions []ActionInfo
	Subject SubjectInfo
	Object  ObjectInfo
}

type ActionInfo struct {
	URI string
}

type SubjectInfo struct {
	AuthenticatedUsers []string
}

type ObjectInfo struct {
	Resources []string
}
