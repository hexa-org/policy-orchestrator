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
    input.principals[_] in policy.subject.authenticated_users
}
matches(action) {
    input.method == action.uri
}
