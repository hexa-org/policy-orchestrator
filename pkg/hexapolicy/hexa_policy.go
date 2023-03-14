package hexapolicy

const (
	SAnyUser   string = "any"
	SAnyAuth   string = "anyAuthenticated"
	SBasicAuth string = "basic"
	SJwtAuth   string = "jwt"
	SSamlAuth  string = "saml"
	SCidr      string = "net"
)

type Policies struct {
	Policies []PolicyInfo `json:"policies"`
}

func (p *Policies) AddPolicy(info PolicyInfo) {
	p.Policies = append(p.Policies, info)
}

func (p *Policies) AddPolicies(policies Policies) {
	for _, v := range policies.Policies {
		p.AddPolicy(v)
	}
}

type PolicyInfo struct {
	Meta      MetaInfo       `validate:"required"`
	Subject   SubjectInfo    `validate:"required"`
	Actions   []ActionInfo   `validate:"required"`
	Object    ObjectInfo     `validate:"required"`
	Condition *ConditionInfo `json:",omitempty"` // Condition is optional
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
	ResourceID string `json:"resource_id" validate:"required"`
}
