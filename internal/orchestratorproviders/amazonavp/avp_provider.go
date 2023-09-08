package amazonavp

import (
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hexa-org/policy-mapper/hexaIdql/pkg/hexapolicy"
	"github.com/hexa-org/policy-mapper/mapper/formats/awsCedar"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonavp/avpClient"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
)

type AvpMeta struct {
	PolicyId   *string
	StoreId    *string
	PolicyType string
	Principal  interface{}
	Resource   interface{}
}

func MapAvpMeta(item types.PolicyItem) AvpMeta {
	return AvpMeta{
		PolicyId:   item.PolicyId,
		StoreId:    item.PolicyStoreId,
		Principal:  item.Principal,
		Resource:   item.Resource,
		PolicyType: string(types.PolicyTypeStatic),
	}
}

func MapAvpTemplate(item *verifiedpermissions.GetPolicyTemplateOutput) AvpMeta {
	return AvpMeta{
		PolicyId:   item.PolicyTemplateId,
		StoreId:    item.PolicyStoreId,
		PolicyType: string(types.PolicyTypeTemplateLinked),
	}
}

type AmazonAvpProvider struct {
	AwsClientOpts awscommon.AWSClientOptions
	cedarMapper   *awsCedar.CedarPolicyMapper
}

func (a *AmazonAvpProvider) Name() string {
	return "avp"
}

func (a *AmazonAvpProvider) initCedarMapper() {
	if a.cedarMapper == nil {
		a.cedarMapper = awsCedar.New(map[string]string{})
	}
}

func (a *AmazonAvpProvider) getAvpClient(info orchestrator.IntegrationInfo) (avpClient.AvpClient, error) {
	var err error
	client, err := avpClient.NewAvpClient(info.Key, a.AwsClientOpts) // NewFromConfig(info.Key, a.AwsClientOpts)
	if err != nil {
		return nil, err
	}
	a.initCedarMapper()

	return client, nil
}

func (a *AmazonAvpProvider) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	if !strings.EqualFold(info.Name, a.Name()) {
		return []orchestrator.ApplicationInfo{}, nil
	}

	client, err := a.getAvpClient(info)
	if err != nil {
		return nil, err
	}

	return client.ListStores()
}

func (a *AmazonAvpProvider) GetPolicyInfo(info orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo) ([]hexapolicy.PolicyInfo, error) {
	client, err := a.getAvpClient(info)
	if err != nil {
		return nil, err
	}
	hexaPols := make([]hexapolicy.PolicyInfo, 0)

	avpPolicies, err := client.ListPolicies(applicationInfo)
	if err != nil {
		return nil, err
	}

	for _, avpPolicy := range avpPolicies {
		policyType := avpPolicy.PolicyType

		switch policyType {
		case types.PolicyTypeStatic:
			output, err := client.GetPolicy(avpPolicy.PolicyId, applicationInfo)
			if err != nil {
				return nil, err
			}
			policyDefinition := output.Definition
			policyStatic := policyDefinition.(*types.PolicyDefinitionDetailMemberStatic).Value
			cedarPolicy := policyStatic.Statement
			mapPols, err := a.cedarMapper.ParseAndMapCedarToHexa([]byte(*cedarPolicy))
			if err != nil {
				return nil, err
			}
			hexaPolicy := mapPols.Policies[0]

			// Update IDQL Meta
			avpMeta := MapAvpMeta(avpPolicy)
			hexaPolicy.Meta.SourceMeta = avpMeta
			hexaPolicy.Meta.Created = avpPolicy.CreatedDate
			hexaPolicy.Meta.Modified = avpPolicy.LastUpdatedDate
			hexaPolicy.Meta.Description = *policyStatic.Description

			hexaPols = append(hexaPols, hexaPolicy)

		case types.PolicyTypeTemplateLinked:
			continue
			// TODO: Pending bug fix in policy-mapper
			/*
				policyDefinition := avpPolicy.Definition
				policyLinked := policyDefinition.(*types.PolicyDefinitionItemMemberTemplateLinked).Value

				output, err := client.GetTemplatePolicy(policyLinked.PolicyTemplateId, applicationInfo)
				if err != nil {
					return nil, err
				}
				// permit(
				//    principal == ?principal,
				//    action in [hexa_avp::Action::"ReadAccount"],
				//    resource == ?resource
				// );
				policyString := *output.Statement
				mapPols, err := a.cedarMapper.ParseAndMapCedarToHexa([]byte(policyString))
				if err != nil {
					return nil, err
				}
				hexaPolicy := mapPols.Policies[0]

				// Update the meta information
				avpMeta := MapAvpTemplate(output)
				hexaPolicy.Meta.SourceMeta = avpMeta
				if output.Description != nil {
					hexaPolicy.Meta.Description = *output.Description
					hexaPolicy.Meta.Created = output.CreatedDate
					hexaPolicy.Meta.Modified = output.LastUpdatedDate
				}
				hexaPols = append(hexaPols, hexaPolicy)

			*/

		default:
			continue
		}

	}
	// Now to map the policies
	return hexaPols, nil
}

func (a *AmazonAvpProvider) SetPolicyInfo(info orchestrator.IntegrationInfo, applicationInfo orchestrator.ApplicationInfo, hexaPolicies []hexapolicy.PolicyInfo) (int, error) {
	client, err := a.getAvpClient(info)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	// Get all existing policies to compare:
	avpExistingPolicies, err := client.ListPolicies(applicationInfo)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var avpMap map[string]types.PolicyItem
	for _, policy := range avpExistingPolicies {
		avpMap[*policy.PolicyId] = policy
	}

	for _, hexaPolicy := range hexaPolicies {
		cedarPolicies, err := a.cedarMapper.MapPolicyToCedar(hexaPolicy)
		if err != nil {
			return http.StatusBadRequest, err
		}

		var cedarDefinition string
		for i, cedarPolicy := range cedarPolicies {
			if i != 0 {
				cedarDefinition = cedarDefinition + "\n"
			}
			cedarDefinition = cedarDefinition + cedarPolicy.String()
		}

		updatePolicyDefinition := types.UpdateStaticPolicyDefinition{
			Statement:   &cedarDefinition,
			Description: &hexaPolicy.Meta.Description,
		}

		updateMemberStatic := types.UpdatePolicyDefinitionMemberStatic{Value: updatePolicyDefinition}

		if hexaPolicy.Meta.SourceMeta != nil {
			// Can assume this is an update
			client.UpdatePolicy(&verifiedpermissions.UpdatePolicyInput{
				Definition:    &updateMemberStatic,
				PolicyId:      nil,
				PolicyStoreId: nil,
			})
		}
	}
	// For each policy, check the Meta.SourceMeta structure to see if this is was a pre-existing policy
	return http.StatusNotImplemented, nil
}
