package policysupport // todo - longer name used here to simplify a refactoring

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
