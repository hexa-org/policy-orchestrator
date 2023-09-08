package hexapolicy

// todo - longer name used here to simplify a refactoring

type PolicyInfo struct {
	Meta    MetaInfo     `validate:"required"`
	Actions []ActionInfo `validate:"required"`
	Subject SubjectInfo  `validate:"required"`
	Object  ObjectInfo   `validate:"required"`
}

type MetaInfo struct {
	Version string `validate:"required"`
}

type ActionInfo struct {
	ActionUri string `validate:"required"`
}

type SubjectInfo struct {
	Members []string `validate:"required"`
}

type ObjectInfo struct {
	ResourceID string `validate:"required"`
}
