package open_policy_agent

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"hexa/pkg/adapters"
	"html/template"
	"io"
	"path/filepath"
	"strings"
)

type OpaAdapter struct {
	resourcesDirectory string
}

func NewOpaAdapter(resourcesDirectory string) adapters.Adapter {
	return OpaAdapter{resourcesDirectory}
}

func (o OpaAdapter) WritePolicies(policies []adapters.Policy, destination io.Writer) error {
	templates := []string{filepath.Join(o.resourcesDirectory, fmt.Sprintf("./template.gohtml"))}
	err := template.Must(template.ParseFiles(templates...)).Execute(destination, policies)
	if err != nil {
		return err
	}
	return nil
}

func (o OpaAdapter) ReadPolicies(source io.Reader) ([]adapters.Policy, error) {
	var rego []byte
	rego, err := io.ReadAll(source)
	if err != nil {
		return []adapters.Policy{}, err
	}

	ast := &Rego{}
	parser, _ := participle.Build(&Rego{}, participle.Unquote())
	err = parser.ParseString("policy.rego", string(rego), ast)
	if err != nil {
		return []adapters.Policy{}, err
	}

	var policies []adapters.Policy
	for _, policy := range ast.Policies {

		var resources []string
		for _, resource := range policy.Info.Path.Array {
			resources = append(resources, resource.String())
		}

		var principals []string
		for _, principal := range policy.Info.Principal.Array {
			principals = append(principals, principal.String())
		}

		found := adapters.Policy{
			Action:  policy.Info.Method.String(),
			Object:  adapters.Object{Resources: resources},
			Subject: adapters.Subject{AuthenticatedUsers: principals},
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
		return fmt.Sprintf("%s", *l.Str)
	case l.Array != nil:
		out := []string{}
		for _, v := range l.Array {
			out = append(out, v.String())
		}
		return fmt.Sprintf("[]*Value{ %s }", strings.Join(out, ", "))
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
