package hexapolicysupport

import (
	"encoding/json"
	"os"

	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
)

// ParsePolicyFile parses a file containing IDQL policy data in JSON form. The top level attribute is "policies" which
// is an array of IDQL Policies ([]PolicyInfo)
func ParsePolicyFile(path string) ([]hexapolicy.PolicyInfo, error) {
	policyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParsePolicies(policyBytes)
}

// ParsePolicies parses an array of bytes representing an IDQL policy data in JSON form. The top level attribute is "policies" which
// is an array of IDQL Policies ([]PolicyInfo)
func ParsePolicies(policyBytes []byte) ([]hexapolicy.PolicyInfo, error) {
	var policies hexapolicy.Policies
	err := json.Unmarshal(policyBytes, &policies)
	if err != nil {
		// Try array of polcies
		var pols []hexapolicy.PolicyInfo
		err = json.Unmarshal(policyBytes, &pols)
		if err != nil {
			return nil, err
		}
		return pols, nil
	}
	return policies.Policies, nil
}

func ToBytes(policies []hexapolicy.PolicyInfo) ([]byte, error) {
	pol := hexapolicy.Policies{Policies: policies}
	return json.Marshal(&pol)
}

func WritePolicies(path string, policies []hexapolicy.PolicyInfo) error {
	polBytes, err := ToBytes(policies)
	if err != nil {
		return err
	}
	return os.WriteFile(path, polBytes, 0644)
}
