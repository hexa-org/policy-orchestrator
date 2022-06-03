package policysupport

// todo - longer name used here to simplify a refactoring

type PolicyInfo struct {
	Meta    MetaInfo
	Actions []ActionInfo
	Subject SubjectInfo
	Object  ObjectInfo
}

type MetaInfo struct {
	Version string
}

type ActionInfo struct {
	Action string
}

type SubjectInfo struct {
	AuthenticatedUsers []string
}

type ObjectInfo struct {
	Resources []string
}
