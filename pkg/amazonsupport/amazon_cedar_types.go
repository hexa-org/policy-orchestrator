package amazonsupport

/*
import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

const (
	PERMIT string = "permit"

	IN string = "in"

	EQUALS string = "=="
	WHEN   string = "when"
	UNLESS string = "unless"
	TERM   string = ";"

	SPACER string = "    "
)

type CedarPolicies struct {
	Policies []*CedarPolicy `parser:"(@@ ';')+"`
}

type CedarPolicy struct {
	Type       string               `parser:"@('permit'|'forbid')"`
	Head       *PolicyHead          `parser:"'(' @@ ')'"`
	Conditions []*ConditionalClause `parser:"@@*"`
}

func (c *CedarPolicy) String() string {
	doc := string(c.Type) + " "
	doc = doc + c.Head.String()
	if c.Conditions != nil {
		for _, v := range c.Conditions {
			doc = doc + "\n" + v.String()

		}
	}
	doc = doc + TERM + "\n"
	return doc
}

type PolicyHead struct {
	Principal *PrincipalExpression `parser:"'principal' @@? ','"` // ser:"'principal' @@? ','"`
	Actions   *ActionExpression    `parser:"'action' @@? ','"`
	Resource  *ResourceExpression  `parser:"'resource' @@?"` // `parser:"'resource' @@? "`
}

func getType(entity string) string {
	parts := strings.SplitN(entity, "::", 2)
	if len(parts) < 2 {
		parts = strings.SplitN(entity, ":", 2)
	}
	return parts[0]
}

func (p *PolicyHead) String() string {
	doc := "(\n"
	if p.Principal == nil {
		doc = doc + SPACER + "principal,\n"
	} else {
		pType := getType(p.Principal.Entity)
		if isSingular(pType) {
			doc = doc + SPACER + "principal == " + p.Principal.Entity + ",\n"
		} else {
			doc = doc + SPACER + "principal in " + p.Principal.Entity + ",\n"
		}

	}

	if p.Actions == nil {
		doc = doc + SPACER + "action,\n"
	} else {
		doc = doc + SPACER + "action " + p.Actions.String() + ",\n"
	}

	if p.Resource == nil {
		doc = doc + SPACER + "resource\n"
	} else {
		rType := getType(p.Resource.Entity)
		if isSingular(rType) {
			doc = doc + SPACER + "resource == " + p.Resource.Entity + "\n"
		} else {
			doc = doc + SPACER + "resource in " + p.Resource.Entity + "\n"
		}
	}
	return doc + ")"
}

type ConditionType string

func (c *ConditionType) Parse(lex *lexer.PeekingLexer) error {
	buf := strings.Builder{}
	isDouble := false
	tok := lex.RawPeek()
	for {
		if !isDouble {
			tok = lex.RawPeek()
		} else {
			isDouble = false
		}
		if tok.Value == "{" {
			lex.Next()
			tok = lex.Peek()
		}

		if tok.EOF() {
			break
		}
		val := tok.String()

		if val == "}" {
			lex.Next()
			break
		}

		switch val {
		case "in":
			buf.WriteString(" " + val + " ")
		case "&", "|", ",", "<", ">", "=":
			lex.Next()
			tok = lex.RawPeek()
			switch tok.String() {
			case "&", "|", ",", "<", ">", "=":
				buf.WriteString(" " + val + tok.String() + " ")
			default:
				buf.WriteString(" " + val + " ")
				isDouble = true
			}

		default:
			buf.WriteString(val)
		}
		if !isDouble {
			lex.Next()
		}

	}
	expr := buf.String()
	// fmt.Println(expr)
	*c = ConditionType(expr)
	return nil
}

type ConditionalClause struct {
	Type string `parser:"@('when'|'unless')"`
	// Condition string `parser:"'{'@(Ident|String|' '|':'|'.'|'='|'&')+'}'"`
	// Condition string `parser:"'{' @(~'}' ' '*)+ '}'"`
	Condition *ConditionType `parser:"@@"`
}

func (c *ConditionalClause) String() string {
	cond := string(*c.Condition)
	if c.Type == WHEN {
		return "when { " + cond + " }"
	}
	return "unless { " + cond + " }"
}

type PrincipalExpression struct {
	Operator string `parser:"@('=''='|'in'|'IN')"` // `@("=" "="|"in"|"IN")`
	// Operator string `parser:"@('=='|'in'|'IN')"`
	Entity string `parser:"@(Ident|':'|String)+"`
}

func (e *PrincipalExpression) String() string {
	if e.Operator == EQUALS {
		return "== " + e.Entity
	}
	return "in [" + e.Entity + "]"
}

type ResourceExpression struct {
	Operator string `parser:"@('=''='|'in'|'IN')"`
	// Operator string `parser:"@('=='|'in'|'IN')"`
	Entity string `parser:"@(Ident|':'|String)+"`
}

func (e *ResourceExpression) String() string {
	if e.Operator == EQUALS {
		return "== " + e.Entity
	}
	return "in [" + e.Entity + "]"
}

type ActionItem struct {
	Item string `parser:"@(Ident|':'|String)+"`
}

type ActionExpression struct {
	Operator string       `parser:"@('=''='|'in')"` // @("=" "="|"in"|"IN")`
	Actions  []ActionItem `parser:"('[' (@@ ','? )* ']')?"`
	Action   string       `parser:"(@(Ident|':'|String)+)?"`
}

func (a *ActionExpression) String() string {

	if a.Operator == EQUALS {
		return "== " + a.Action
	}

	listString := ""
	for k, v := range a.Actions {
		if k == 0 {
			listString = v.Item

		} else {
			listString = listString + ", " + v.Item
		}
	}
	return "in [" + listString + "]"
}
*/
