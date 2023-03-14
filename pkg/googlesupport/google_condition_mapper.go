package googlesupport

/*
 Condition mapper for Google IAM - See: https://cloud.google.com/iam/docs/conditions-overview
*/
import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	celv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/filters/cel/v3"
	"github.com/google/cel-go/cel"
	"github.com/hexa-org/policy-orchestrator/pkg/filtersupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var (
	env, _ = cel.NewEnv()
)

type GoogleConditionMapper struct {
	NameMapper *hexapolicy.AttributeMap
}

type CelConditionType struct {
	Title       string
	Description string
	Expression  celv3.ExpressionFilter
}

type GcpConditionType struct {
	Title       string
	Description string
	Expression  string
}

func (mapper *GoogleConditionMapper) MapConditionToProvider(condition hexapolicy.ConditionInfo) (string, error) {
	// assumes https://github.com/google/cel-spec/blob/master/doc/langdef.md#logical-operators
	ast, err := hexapolicy.ParseConditionRuleAst(condition)
	if err != nil {
		return "", err
	}
	return mapper.MapFilter(ast)

}

func (mapper *GoogleConditionMapper) MapFilter(ast *filtersupport.Expression) (string, error) {
	err := checkCompatibility(*ast)
	if err != nil {
		return "", err
	}
	return mapper.mapFilterInternal(ast, false), nil
}

func (mapper *GoogleConditionMapper) mapFilterInternal(ast *filtersupport.Expression, isChild bool) string {

	// dereference
	deref := *ast

	switch element := deref.(type) {
	case filtersupport.NotExpression:
		return mapper.mapFilterNot(&element, isChild)
	case filtersupport.PrecedenceExpression:
		return mapper.mapFilterPrecedence(&element, true)

	case filtersupport.LogicalExpression:
		return mapper.mapFilterLogical(&element, isChild)

	default:
		attrExpression := deref.(filtersupport.AttributeExpression)
		return mapper.mapFilterAttrExpr(&attrExpression)
	}
	// return mapper.mapFilterValuePath(deref.(filter.ValuePathExpression))
}

/*
func (mapper *GoogleConditionMapper) mapFilterValuePath(vpFilter filter.ValuePathExpression) string {
	// See: https://cloud.google.com/access-context-manager/docs/custom-access-level-spec
	subFilter := vpFilter.VPathFilter
	attribute := vpFilter.Attribute
	celFilter := mapper.mapFilterInternal(&subFilter, false)
	return attribute + ".exists(" + attribute + "," + celFilter + ")"
}
*/

func (mapper *GoogleConditionMapper) mapFilterNot(notFilter *filtersupport.NotExpression, isChild bool) string {
	subExpression := notFilter.Expression
	var celFilter string
	switch subFilter := subExpression.(type) {
	case filtersupport.LogicalExpression:
		// For the purpose of a not filter, the logical expression is not a child
		celFilter = mapper.mapFilterLogical(&subFilter, false)
		celFilter = "(" + celFilter + ")"
		break
	default:
		celFilter = mapper.mapFilterInternal(&subFilter, false)
	}

	return fmt.Sprintf("!%v", celFilter)
}

func (mapper *GoogleConditionMapper) mapFilterPrecedence(pfilter *filtersupport.PrecedenceExpression, isChild bool) string {
	subExpression := pfilter.Expression
	var celFilter string
	switch subFilter := subExpression.(type) {
	case filtersupport.LogicalExpression:
		// For the purpose of a not filter, the logical expression is not a child
		celFilter = mapper.mapFilterLogical(&subFilter, false)
		celFilter = "(" + celFilter + ")"
		break
	default:
		celFilter = mapper.mapFilterInternal(&subFilter, false)
	}
	return fmt.Sprintf("%v", celFilter)
}

func (mapper *GoogleConditionMapper) mapFilterLogical(logicFilter *filtersupport.LogicalExpression, isChild bool) string {
	isDouble := false
	var celLeft, celRight string
	switch subFilter := logicFilter.Left.(type) {
	case filtersupport.LogicalExpression:
		if subFilter.Operator == logicFilter.Operator {
			isDouble = true
		}
	}

	celLeft = mapper.mapFilterInternal(&logicFilter.Left, !isDouble)

	celRight = mapper.mapFilterInternal(&logicFilter.Right, !isDouble)

	switch logicFilter.Operator {
	default:
		return fmt.Sprintf("%v && %v", celLeft, celRight)
	case filtersupport.OR:
		if isChild {
			// Add precedence to preserve order
			return fmt.Sprintf("(%v || %v)", celLeft, celRight)
		} else {
			return fmt.Sprintf("%v || %v", celLeft, celRight)
		}
	}
}

func (mapper *GoogleConditionMapper) mapFilterAttrExpr(attrExpr *filtersupport.AttributeExpression) string {
	compareValue := prepareValue(attrExpr)

	mapPath := mapper.NameMapper.GetProviderAttributeName(attrExpr.AttributePath)

	switch attrExpr.Operator {

	case filtersupport.NE:
		return mapPath + " != " + compareValue
	case filtersupport.LT:
		return mapPath + " < " + compareValue
	case filtersupport.LE:
		return mapPath + " <= " + compareValue
	case filtersupport.GT:
		return mapPath + " > " + compareValue
	case filtersupport.GE:
		return mapPath + " >= " + compareValue
	case filtersupport.SW:
		return mapPath + ".startsWith(" + compareValue + ")"
	case filtersupport.EW:
		return mapPath + ".endsWith(" + compareValue + ")"
	case filtersupport.PR:
		return "has(" + mapPath + ")"
	case filtersupport.CO:
		return mapPath + ".contains(" + compareValue + ")"
	case filtersupport.IN:
		return mapPath + " in " + compareValue
	default:
		return mapPath + " == " + compareValue
	}

}

/*
If the value type is string, it needs to be quoted.
*/
func prepareValue(attrExpr *filtersupport.AttributeExpression) string {
	compValue := attrExpr.CompareValue
	if compValue == "" {
		return ""
	}

	// Check for integer

	if _, err := strconv.ParseInt(compValue, 10, 64); err == nil {
		return compValue
	}

	if compValue == "true" || compValue == "false" {
		return compValue
	}

	// assume it is a string and return with quotes
	return fmt.Sprintf("\"%s\"", attrExpr.CompareValue)

}

func (mapper *GoogleConditionMapper) MapProviderToCondition(expression string) (hexapolicy.ConditionInfo, error) {

	celAst, issues := env.Parse(expression)
	if issues != nil {
		return hexapolicy.ConditionInfo{}, errors.New("CEL Mapping Error: " + issues.String())
	}

	idqlAst, err := mapper.mapCelExpr(celAst.Expr(), false)
	if err != nil {
		return hexapolicy.ConditionInfo{
			Rule: "",
		}, errors.New("IDQL condition mapper error: " + err.Error())
	}

	condString := hexapolicy.SerializeExpression(&idqlAst)

	return hexapolicy.ConditionInfo{
		Rule:   condString,
		Action: "allow",
	}, nil
}

func (mapper *GoogleConditionMapper) mapCelExpr(expression *expr.Expr, isChild bool) (filtersupport.Expression, error) {

	cexpr := expression.GetCallExpr()

	if cexpr != nil {
		return mapper.mapCallExpr(cexpr, isChild)
	}

	kind := expression.GetExprKind()
	switch v := kind.(type) {
	case *expr.Expr_SelectExpr:
		return mapper.mapSelectExpr(v)
	// case *expr.Expr_ComprehensionExpr:
	//	return nil, errors.New("unimplemented CEL 'comprehension expression' not implemented. ")
	default:
		msg := fmt.Sprintf("unimplemented CEL expression: %s", expression.String())
		return nil, fmt.Errorf(msg)
	}
}

func (mapper *GoogleConditionMapper) mapSelectExpr(selection *expr.Expr_SelectExpr) (filtersupport.Expression, error) {
	field := selection.SelectExpr.GetField()
	/*
		if !selection.SelectExpr.GetTestOnly() {
			return nil, errors.New("unimplemented Google CEL Select Expression: " + selection.SelectExpr.String())
		}
	*/

	ident := selection.SelectExpr.GetOperand().GetIdentExpr()

	name := ident.GetName()
	attr := name + "." + field
	path := mapper.NameMapper.GetHexaFilterAttributePath(attr)
	return filtersupport.AttributeExpression{
		AttributePath: path,
		Operator:      filtersupport.PR,
	}, nil
}

func (mapper *GoogleConditionMapper) mapCallExpr(expression *expr.Expr_Call, isChild bool) (filtersupport.Expression, error) {
	operand := expression.GetFunction()
	switch operand {
	case "_&&_":
		return mapper.mapCelLogical(expression.Args, true, isChild)
	case "_||_":
		return mapper.mapCelLogical(expression.Args, false, isChild)
	case "_!_", "!_":
		return mapper.mapCelNot(expression.Args, isChild), nil // was false
	case "_==_":
		return mapper.mapCelAttrCompare(expression.Args, filtersupport.EQ)
	case "_!=_":
		return mapper.mapCelAttrCompare(expression.Args, filtersupport.NE)
	case "_>_":
		return mapper.mapCelAttrCompare(expression.Args, filtersupport.GT)
	case "_<_":
		return mapper.mapCelAttrCompare(expression.Args, filtersupport.LT)
	case "_<=_":
		return mapper.mapCelAttrCompare(expression.Args, filtersupport.LE)
	case "_>=_":
		return mapper.mapCelAttrCompare(expression.Args, filtersupport.GE)
	case "@in":
		return mapper.mapCelAttrCompare(expression.Args, filtersupport.IN)

	case "startsWith", "endsWith", "contains", "has":
		return mapper.mapCelAttrFunction(expression)

	}

	return nil, errors.New("unimplemented CEL expression operand: " + operand)
}

func (mapper *GoogleConditionMapper) mapCelAttrFunction(expression *expr.Expr_Call) (filtersupport.Expression, error) {
	target := expression.GetTarget()
	selection := target.GetSelectExpr()
	operand := selection.GetOperand()
	// subattr := selection.Field
	name := operand.GetIdentExpr().GetName()
	var path string
	if name == "" {
		name = target.GetIdentExpr().GetName()
		path = mapper.NameMapper.GetHexaFilterAttributePath(name)

	} else {
		subattr := selection.GetField()
		path = mapper.NameMapper.GetHexaFilterAttributePath(name + "." + subattr)
	}

	switch expression.GetFunction() {
	case "startsWith":
		rh := expression.GetArgs()[0].GetConstExpr().GetStringValue()
		return filtersupport.AttributeExpression{
			AttributePath: path,
			Operator:      filtersupport.SW,
			CompareValue:  rh,
		}, nil
	case "endsWith":
		rh := expression.GetArgs()[0].GetConstExpr().GetStringValue()
		return filtersupport.AttributeExpression{
			AttributePath: path,
			Operator:      filtersupport.EW,
			CompareValue:  rh,
		}, nil
	case "contains":
		rh := expression.GetArgs()[0].GetConstExpr().GetStringValue()
		return filtersupport.AttributeExpression{
			AttributePath: path,
			Operator:      filtersupport.CO,
			CompareValue:  rh,
		}, nil
	}
	return nil, errors.New(fmt.Sprintf("unimplemented CEL function:%s", expression.GetFunction()))

}

func (mapper *GoogleConditionMapper) mapCelAttrCompare(expressions []*expr.Expr, operator filtersupport.CompareOperator) (filtersupport.Expression, error) {
	// target :=

	path := ""
	isNot := false
	callExpr := expressions[0].GetCallExpr()
	lhExpression := expressions[0]
	if callExpr != nil {
		switch callExpr.GetFunction() {
		case "!_":
			isNot = true
			break
		default:
			msg := fmt.Sprintf("unimplemented CEL function: %s", callExpr.GetFunction())
			return nil, errors.New(msg)
		}
		lhExpression = callExpr.Args[0]
	}
	ident := lhExpression.GetIdentExpr()
	if ident == nil {
		selectExpr := lhExpression.GetSelectExpr()
		path = selectExpr.GetOperand().GetIdentExpr().Name + "." + selectExpr.GetField()
	} else {
		path = ident.GetName()
	}

	// map the path name
	path = mapper.NameMapper.GetHexaFilterAttributePath(path)
	constExpr := expressions[1].GetConstExpr().String()

	elems := strings.SplitN(constExpr, ":", 2)
	rh := ""

	if len(elems) == 2 {
		switch elems[0] {
		case "string_value":
			rh = expressions[1].GetConstExpr().GetStringValue()
		default:
			rh = elems[1]
		}
	}
	attrFilter := filtersupport.AttributeExpression{
		AttributePath: path,
		Operator:      operator,
		CompareValue:  rh,
	}
	if isNot {
		return filtersupport.NotExpression{
			Expression: attrFilter,
		}, nil
	}
	return attrFilter, nil
}

func (mapper *GoogleConditionMapper) mapCelNot(expressions []*expr.Expr, isChild bool) filtersupport.Expression {

	expression, _ := mapper.mapCelExpr(expressions[0], false) // ischild is ignored because of not

	notFilter := filtersupport.NotExpression{
		Expression: expression,
	}
	return notFilter
}

func (mapper *GoogleConditionMapper) mapCelLogical(expressions []*expr.Expr, isAnd bool, isChild bool) (filtersupport.Expression, error) {
	filters := make([]filtersupport.Expression, len(expressions))
	var err error
	// collapse n clauses back into a series of nested pairwise and/or clauses
	for i, v := range expressions {
		filters[i], err = mapper.mapCelExpr(v, true)
		if err != nil {
			return nil, err
		}
	}
	var op string
	if isAnd {
		op = "and"
	} else {
		op = "or"

	}

	// Collapse all the way down to 1 filter
	for len(filters) > 1 {
		i := len(filters)
		subFilter := filtersupport.LogicalExpression{
			Left:     filters[i-2],
			Right:    filters[i-1],
			Operator: filtersupport.LogicalOperator(op),
		}

		filters[i-2] = subFilter
		filters = filters[0 : i-1 : i-1]
	}

	// Surround with precedence to preserve order
	return filters[0], nil
}

func checkCompatibility(e filtersupport.Expression) error {
	var err error
	switch v := e.(type) {
	case filtersupport.LogicalExpression:
		err = checkCompatibility(v.Left)
		if err != nil {
			return err
		}
		err = checkCompatibility(v.Right)
		if err != nil {
			return err
		}
	case filtersupport.NotExpression:
		return checkCompatibility(v.Expression)
	case filtersupport.PrecedenceExpression:
		return checkCompatibility(v.Expression)
	case filtersupport.ValuePathExpression:
		return errors.New("IDQL ValuePath expression mapping to Google CEL currently not supported")
	case filtersupport.AttributeExpression:
		return nil
	}
	return nil
}
