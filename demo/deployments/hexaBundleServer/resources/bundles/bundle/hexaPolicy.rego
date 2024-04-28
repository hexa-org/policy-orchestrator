package hexaPolicy

# Rego Policy Interpreter for IDQL V0.62.1b (IDQL)
import rego.v1

import data.policies

# Returns whether the current operation is allowed
allow if {
	count(allowSet) > 0
}

# Returns the list of possible actions allowed (e.g. for UI buttons)
actionRights contains name if {
	some policy in policies
	policy.meta.policyId in allowSet

	some action in policy.actions
	name := sprintf("%s/%s", [policy.meta.policyId, action.actionUri])
}

# Returns the list of matching policy names based on current request
allowSet contains name if {
	some policy in policies
	subjectMatch(policy.subject, input.subject, input.req)
	actionsMatch(policy.actions, input.req)
	objectMatch(policy.object, input.req)
	conditionMatch(policy, input)

	name := policy.meta.policyId # this will be id of the policy
}

subjectMatch(psubject, _, _) if {
	# Match if no members value specified
	not psubject.members
}

subjectMatch(psubject, insubject, req) if {
	# Match if no members value specified
	some member in psubject.members
	subjectMemberMatch(member, insubject, req)
}

subjectMemberMatch(member, _, _) if {
	# If policy is any that we will skip processing of subject
	lower(member) == "any"
}

subjectMemberMatch(member, insubj, _) if {
	# anyAutheticated - A match occurs if input.subject has a value other than anonymous and exists.
	insubj.sub # check sub exists
	lower(member) == "anyauthenticated"
}

# Check for match based on user:<sub>
subjectMemberMatch(member, insubj, _) if {
	startswith(lower(member), "user:")
	user := substring(member, 5, -1)
	lower(user) == lower(insubj.sub)
}

# Check for match if sub ends with domain
subjectMemberMatch(member, insubj, _) if {
	startswith(lower(member), "domain:")
	domain := lower(substring(member, 7, -1))
	endswith(lower(insubj.sub), domain)
}

# Check for match based on role
subjectMemberMatch(member, insubj, _) if {
	startswith(lower(member), "role:")
	role := substring(member, 5, -1)
	role in insubj.roles
}

subjectMemberMatch(member, _, req) if {
    startswith(lower(member), "net:")
	cidr := substring(member, 4, -1)
	addr := split(req.ip, ":")  # Split because IP is address:port
	net.cidr_contains(cidr, addr[0])
}

actionsMatch(actions, _) if {
	# no actions is a match
	not actions
}

actionsMatch(actions, req) if {
	some action in actions
	actionMatch(action, req)
}

actionMatch(action, req) if {
	# Check for match based on ietf http
	checkIetfMatch(action.actionUri, req)
}

actionMatch(action, req) if {
	action.actionUri # check for an action
	count(req.actionUris) > 0

	# Check for a match based on req.ActionUris and actionUri
	checkUrnMatch(action.actionUri, req.actionUris)
}

checkUrnMatch(policyUri, actionUris) if {
	some action in actionUris
	lower(policyUri) == lower(action)
}

checkIetfMatch(actionUri, req) if {
	# first match the rule against literals
	components := split(lower(actionUri), ":")
	count(components) > 2
	components[0] == "ietf"
	startswith(components[1], "http")

	startswith(lower(input.req.protocol), "http")
	checkHttpMethod(components[2], req.method)

	checkPath(components[3], req)
}

objectMatch(object, req) if {
    not object
    not object.resource_id
}

objectMatch(object, req) if {
	object.resource_id

	some reqUri in req.resourceIds
    lower(object.resource_id) == lower(reqUri)
}

checkHttpMethod(allowMask, _) if {
	contains(allowMask, "*")
}

checkHttpMethod(allowMask, reqMethod) if {
	startswith(allowMask, "!")

	not contains(allowMask, lower(reqMethod))
}

checkHttpMethod(allowMask, reqMethod) if {
	not startswith(allowMask, "!")
	contains(allowMask, lower(reqMethod))
}

checkPath(path, req) if {
	path # if path specified it must match
	glob.match(path, ["*"], req.path)
}

checkPath(path, _) if {
	not path # if path not specified, it will not be matched
}

conditionMatch(policy, _) if {
	not policy.condition # Most policies won't have a condition
}

conditionMatch(policy, inreq) if {
	policy.condition
    not policy.condition.action  # Default is to allow
	hexaFilter(policy.condition.rule, inreq) # HexaFilter evaluations the rule for a match against input
}

conditionMatch(policy, inreq) if {
	policy.condition
    action(policy.condition.action)  # if defined, action must be "allow"
	hexaFilter(policy.condition.rule, inreq) # HexaFilter evaluations the rule for a match against input
}

conditionMatch(policy, inreq) if {
    # If action is deny, then hexaFilter must be false
	policy.condition
    not action(policy.condition.action)
	not hexaFilter(policy.condition.rule, inreq) # HexaFilter evaluations the rule for a match against input
}

action(val) if lower(val) == "allow"
