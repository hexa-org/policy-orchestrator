package googlesupport

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
	"google.golang.org/api/iam/v1"
)

type GooglePolicyMapper struct {
	conditionMapper GoogleConditionMapper
}

/*
BindAssignment is an array of GCP Bindings combined with a resource identifier.
*/
type BindAssignment struct {
	ResourceId string        `json:"resource_id"`
	Bindings   []iam.Binding `json:"bindings"`
}

func New(nameMap map[string]string) *GooglePolicyMapper {
	return &GooglePolicyMapper{conditionMapper: GoogleConditionMapper{NameMapper: hexapolicy.NewNameMapper(nameMap)}}
}

func (m *GooglePolicyMapper) Name() string {
	return "bind"
}

func (m *GooglePolicyMapper) MapBindingAssignmentsToPolicy(bindAssignments []*BindAssignment) ([]hexapolicy.PolicyInfo, error) {
	var policies []hexapolicy.PolicyInfo
	for _, v := range bindAssignments {
		pols, err := m.MapBindingAssignmentToPolicy(*v)
		if err != nil {
			return nil, err
		}
		for _, pol := range pols {
			policies = append(policies, pol)
		}
	}
	return policies, nil
}

func (m *GooglePolicyMapper) MapBindingAssignmentToPolicy(bindAssignment BindAssignment) ([]hexapolicy.PolicyInfo, error) {
	var policies []hexapolicy.PolicyInfo
	objectId := bindAssignment.ResourceId
	for _, v := range bindAssignment.Bindings {
		policy, err := m.MapBindingToPolicy(objectId, v)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

func (m *GooglePolicyMapper) MapBindingToPolicy(objectId string, binding iam.Binding) (hexapolicy.PolicyInfo, error) {
	bindingCondition := binding.Condition
	if bindingCondition != nil {
		condition, err := m.convertCelToCondition(binding.Condition)
		if err != nil {
			return hexapolicy.PolicyInfo{}, err
		}

		policy := hexapolicy.PolicyInfo{
			Meta:      hexapolicy.MetaInfo{Version: "0.5"},
			Actions:   convertRoleToAction(binding.Role),
			Subject:   hexapolicy.SubjectInfo{Members: binding.Members},
			Object:    hexapolicy.ObjectInfo{ResourceID: objectId},
			Condition: &condition,
		}
		return policy, nil
	}
	policy := hexapolicy.PolicyInfo{
		Meta:    hexapolicy.MetaInfo{Version: "0.5"},
		Actions: convertRoleToAction(binding.Role),
		Subject: hexapolicy.SubjectInfo{Members: binding.Members},
		Object:  hexapolicy.ObjectInfo{ResourceID: objectId},
	}
	return policy, nil

}

func (m *GooglePolicyMapper) MapPolicyToBinding(policy hexapolicy.PolicyInfo) (*iam.Binding, error) {
	cond := policy.Condition
	var condExpr *iam.Expr
	var err error
	if cond != nil {
		condExpr, err = m.convertPolicyCondition(policy)
	} else {
		condExpr = nil
	}

	if err != nil {
		return nil, err
	}
	return &iam.Binding{
		Condition: condExpr,
		Members:   policy.Subject.Members,
		Role:      convertActionToRole(policy),
	}, nil
}

func (m *GooglePolicyMapper) MapPoliciesToBindings(policies []hexapolicy.PolicyInfo) []*BindAssignment {
	bindingMap := make(map[string][]iam.Binding)

	for i, policy := range policies {
		binding, err := m.MapPolicyToBinding(policy)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		key := policies[i].Object.ResourceID

		existing := bindingMap[key]
		existing = append(existing, *binding)
		bindingMap[key] = existing

	}
	bindings := make([]*BindAssignment, len(bindingMap))
	i := 0
	for k, v := range bindingMap {
		bindings[i] = &BindAssignment{
			ResourceId: k,
			Bindings:   v,
		}
		i++
	}
	return bindings
}

func convertActionToRole(policy hexapolicy.PolicyInfo) string {
	for _, v := range policy.Actions {
		action := v.ActionUri
		if strings.HasPrefix(action, "gcp:") {
			return action[4:]
		}
	}
	return ""
}

func convertRoleToAction(role string) []hexapolicy.ActionInfo {
	if role == "" {
		return nil
	}
	return []hexapolicy.ActionInfo{{"gcp:" + role}}
}

func (m *GooglePolicyMapper) convertCelToCondition(expr *iam.Expr) (hexapolicy.ConditionInfo, error) {
	return m.conditionMapper.MapProviderToCondition(expr.Expression)
}

func (m *GooglePolicyMapper) convertPolicyCondition(policy hexapolicy.PolicyInfo) (*iam.Expr, error) {

	if policy.Condition == nil {
		return nil, nil // do nothing as policy has no condition
	}

	celString, err := m.conditionMapper.MapConditionToProvider(*policy.Condition)
	if err != nil {
		return nil, err
	}

	iamExpr := iam.Expr{
		Expression: celString,
	}
	return &iamExpr, nil
}

type Assignments struct {
	BindAssignments []*BindAssignment
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Assignments) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("no bytes to unmarshal")
	}

	switch b[0] {
	case '{':
		return d.unMarshallSingle(b)
	case '[':
		return d.unMarshallMulti(b)
	}
	return nil
}

func (d *Assignments) unMarshallSingle(b []byte) error {
	type DetectSingle struct {
		iam.Binding
		BindAssignment
	}
	var single DetectSingle
	err := json.Unmarshal(b, &single)
	if err != nil {
		return err
	}
	assignments := make([]*BindAssignment, 1)
	if len(single.Bindings) != 0 {
		assignments[0] = &single.BindAssignment
	} else {
		iamBindings := make([]iam.Binding, 1)
		iamBindings[0] = single.Binding
		assignments[0] = &BindAssignment{
			Bindings:   iamBindings,
			ResourceId: "",
		}
	}
	d.BindAssignments = assignments
	return nil
}

func (d *Assignments) unMarshallMulti(b []byte) error {
	var iamBinds []iam.Binding
	var assigns []BindAssignment

	err := json.Unmarshal(b, &assigns)
	if err != nil {
		return err
	}
	if len(assigns) > 0 {
		pAssigns := make([]*BindAssignment, len(assigns))
		for k := range assigns {
			pAssigns[k] = &assigns[k]
		}
		d.BindAssignments = pAssigns
		return nil
	}

	err = json.Unmarshal(b, &iamBinds)
	if err != nil {
		return err
	}

	assignments := make([]*BindAssignment, 1)
	assignments[0] = &BindAssignment{
		Bindings:   iamBinds,
		ResourceId: "",
	}
	return nil
}

/*
ParseBindings will read either an iam.Binding or GcpBindAssignment structure and returns a []*GcpBindAssignment type.
Note that if a single binding is provided, the GcpBindAssignment.ResourceId value will be nil
*/
func ParseBindings(bindingBytes []byte) ([]*BindAssignment, error) {
	var data Assignments
	err := json.Unmarshal(bindingBytes, &data)
	if err != nil {
		return nil, err
	}

	return data.BindAssignments, nil
}

/*
ParseFile will load a file from the specified path and will auto-detect format and convert to GcpBindAssignment. See ParseBindings
*/
func ParseFile(path string) ([]*BindAssignment, error) {
	policyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBindings(policyBytes)
}
