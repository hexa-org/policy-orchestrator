package authz
import future.keywords.in
default allow = false
allow {
    input.method = "GET"
    input.path in ["/"]
    input.principals[_] in ["allusers"]
}
allow {
    input.method = "GET"
    input.path in ["/sales", "/marketing"]
    input.principals[_] in ["sales@", "marketing@"]
}
allow {
    input.method = "GET"
    input.path in ["/accounting"]
    input.principals[_] in ["accounting@"]
}
allow {
    input.method = "GET"
    input.path in ["/humanresources"]
    input.principals[_] in ["humanresources@"]
}