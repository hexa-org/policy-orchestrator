package openpolicyagent

import (
	"github.com/alecthomas/participle/v2"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"html/template"
	"io"
	"path/filepath"
)

type OpaService struct {
	ResourcesDirectory string
}

func (o OpaService) WritePolicies(policies []provider.PolicyInfo, destination io.Writer) error {
	templates := []string{filepath.Join(o.ResourcesDirectory, "./template.gohtml")}
	must := template.Must(template.ParseFiles(templates...))
	return must.Execute(destination, policies)
}

func (o OpaService) ReadPolicies(source io.Reader) ([]provider.PolicyInfo, error) {
	var rego []byte
	rego, err := io.ReadAll(source)
	if err != nil {
		return []provider.PolicyInfo{}, err
	}

	ast := &Rego{}
	parser, _ := participle.Build(&Rego{}, participle.Unquote())
	err = parser.ParseString("policy.rego", string(rego), ast)
	if err != nil {
		return []provider.PolicyInfo{}, err
	}

	var policies []provider.PolicyInfo
	for _, policy := range ast.Policies {

		var resources []string
		for _, resource := range policy.Info.Path.Array {
			resources = append(resources, resource.String())
		}

		var principals []string
		for _, principal := range policy.Info.Principal.Array {
			principals = append(principals, principal.String())
		}

		found := provider.PolicyInfo{
			Version: "0.4",
			Action:  policy.Info.Method.String(),
			Object:  provider.ObjectInfo{Resources: resources},
			Subject: provider.SubjectInfo{AuthenticatedUsers: principals},
		}
		policies = append(policies, found)
	}
	return policies, nil
}

/// for participle ast below

type Bool bool

func (b *Bool) Capture(v []string) error {
	*b = v[0] == "true"
	return nil
}

type Value struct {
	Str   *string  `@(String|Char|RawString)`
	Array []*Value `| "[" ( @@ ","? )* "]"`
}

func (l *Value) String() string {
	switch {
	case l.Str != nil:
		return *l.Str
	}
	panic("??")
}

type Rego struct {
	Package      *string   `"package" @Ident`
	Import       *string   `"import" "future" "." "keywords" "." "in"`
	DefaultAllow *Bool     `"default" "allow" "=" @("true"|"false")`
	Policies     []*Policy ` @@*`
}

type Policy struct {
	Info *Info `"allow" "{" @@* "}"`
}

type Info struct {
	Method    *Value `"input" "." "method" "=" @@`
	Path      *Value `"input" "." "path" "in" @@`
	Principal *Value `"input" "." "principals" "[" "_" "]" "in" @@`
}
