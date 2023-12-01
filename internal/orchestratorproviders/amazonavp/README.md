
# Amazon AVP API

## ListPolicyStores

Returns a [list of all policy stores](https://docs.aws.amazon.com/verifiedpermissions/latest/apireference/API_ListPolicyStores.html) in the AWS Account.

Request:
```json
{
   "maxResults": number,
   "nextToken": "string"
}
```

Response:
```json
{
   "nextToken": "string",
   "policyStores": [ 
      { 
         "arn": "string",
         "createdDate": "string",
         "policyStoreId": "string"
      }
   ]
}
```

## ListPolicies

Returns a paginated [list of all policies](https://docs.aws.amazon.com/verifiedpermissions/latest/apireference/API_ListPolicies.html) in the specified policy store.

The idea is to be able to get a list of policies that best match a current context (presumably used by isAuthorized)

Request Syntax:
```json lines
{
   "filter": { 
      "policyTemplateId": "string",
      "policyType": "string",
      "principal": { ... },
      "resource": { ... }
   },
   "maxResults": number,
   "nextToken": "string",
   "policyStoreId": "string"
}
```
Note:  nextToken is used to page through results. Each page is maxResults in size.

Response Syntax:
```json lines
{
   "nextToken": "string",
   "policies": [ 
      { 
         "createdDate": "string",
         "definition": { ... },
         "lastUpdatedDate": "string",
         "policyId": "string",
         "policyStoreId": "string",
         "policyType": "string",
         "principal": { 
            "entityId": "string",
            "entityType": "string"
         },
         "resource": { 
            "entityId": "string",
            "entityType": "string"
         }
      }
   ]
}
```

Policies is an array of [PolicyItem](https://docs.aws.amazon.com/verifiedpermissions/latest/apireference/API_PolicyItem.html).

## GetPolicy
Retrieves information about a [specific policy](https://docs.aws.amazon.com/verifiedpermissions/latest/apireference/API_GetPolicy.html). (same as PolicyItem above?)

Request Syntax:
```json lines
{
   "policyId": "string",
   "policyStoreId": "string"
}
```

Response:
```json lines
{
   "createdDate": "string",
   "definition": { ... },
   "lastUpdatedDate": "string",
   "policyId": "string",
   "policyStoreId": "string",
   "policyType": "string",
   "principal": { 
      "entityId": "string",
      "entityType": "string"
   },
   "resource": { 
      "entityId": "string",
      "entityType": "string"
   }
}
```

`definition` is a[ PolicyDefinitionDetail](https://docs.aws.amazon.com/verifiedpermissions/latest/apireference/API_PolicyDefinitionItem.html) object. 
--> this is one of [StaticPolicyDefinition](https://docs.aws.amazon.com/verifiedpermissions/latest/apireference/API_StaticPolicyDefinitionDetail.html) (aka cedar) or TemplateLinkedPolicyDefinition 

A StaticDefinitionDetail consists of:

statement 
: A static policy written in Cedar Policy Language (string)

description
: A description of the policy

## GetPolicyTemplate

A policy template is a policy that contains placeholders. The placeholders can represent the principal and the resource. Later, you can create a template-linked policy based on the policy template by specifying the exact principal and resource to use for this one policy. Template-linked policies are dynamic, meaning that the new policy stays linked to its policy template. When you change a policy statement in the policy template, any policies linked to that template automatically and immediately use the new statement for all authorization decisions made from that moment forward.

You can use placeholders in a Cedar policy template for only the following two elements of a policy statement:

Principal – ?principal
Resource – ?resource
You can use either one or both in a policy template.

Placeholders can appear in only the policy head on the right-hand side of the == or in operators.

Then, when you create a policy based on the policy template, you must specify values for each of the placeholders. Those values are combined with the rest of the policy template to form a complete and usable template-linked policy.

As an example, consider the scenario where a common action is to grant certain groups with the ability to view and comment on any photos that are not marked as private. You decide to associate the action with a Share button in your application’s interface. You could create a template that looks like the following example.

Retrieve a [policy template](https://docs.aws.amazon.com/verifiedpermissions/latest/apireference/API_GetPolicyTemplate.html)...

Request:
```json lines
{
   "policyStoreId": "string",
   "policyTemplateId": "string"
}
```

Response:
```json lines
{
   "createdDate": "string",
   "description": "string",
   "lastUpdatedDate": "string",
   "policyStoreId": "string",
   "policyTemplateId": "string",
   "statement": "string"
}
```

The cedar policy is stored in `statement`.