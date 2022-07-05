package authz
import future.keywords.in
default allow = false
allow {
    anyMatching
}
anyMatching {
    some i
	matches(data.bundle.policies[i])
}
matches(policy) {
    matchesAction(policy.actions[i])
    input.path in policy.object.resources
    matchesPrincipal(policy.subject)
}
matchesAction(action) {
    input.method == action.action_uri
}
matchesPrincipal(subject) {
    "allusers" in subject.members
}
matchesPrincipal(subject) {
    principalExists(input.principal)
    "allauthenticated" in subject.members
}
matchesPrincipal(subject) {
    input.principal in subject.members
}
principalExists(principal) {
    not is_null(principal)
}
