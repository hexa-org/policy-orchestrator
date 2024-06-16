# Security Policy

Hexa Policy Orchestrator follows similar security policy as other CNCF projects.

## Reporting a Vulnerability

Please report any security bugs privately to the maintainers listed in the [MAINTAINERS](MAINTAINERS.md) file. We will
fix the issue and coordinate a release date, acknowledging your effort.

## Overall Hexa Security

Hexa consists of a number of services (see Docker-compose) which work together to demonstrate the control and management
of policy.

### Hexa-Admin

The Hexa Administration application is a web based graphical interface used to administer Hexa and IDQL Policy. When
deployed in a shared environment (outside of development) Hexa-Admin must be deployed behind a secured web proxy in
order to authenticate user sessions and enforce RBAC policy as Hexa-Admin does not implement its own access control
system.

The long-term intent is to leverage authentication providers (e.g. OpenID Connect), establish browser sessions, and
enforce access management using IDQL.

For access to the Hexa-Orchestrator internal API, an OAuth2 Client Credentials flow is used (See [Sec 4.4 of RFC7649](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4)).

### Hexa-Orchestrator

Hexa Orchestrator is the internal API gateway through which policy provisioning and retrieval takes place. Requests from
Hexa-Admin and other container clients are secured using JWT tokens ([RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519)) over TLS.

The current docker-compose demo included with this project configures an OAuth2 service based on [Keycloak](https://keycloak.org).
In the current configuration, the Hexa-Admin client authenticates to the OAuth2 server to obtain JWT tokens to access the 
internal API gateway over TLS.

For demonstration purposes, an initial Keylcoak administrative account is set up and can be found in the `.env_development`
file. Admiistrators should change this password immediately.  Likewise, under the Hexa-Orchestrator-Realm, is the client credential
for `hexaclient`. This secret should be changed and the docker-compose file updated.

When deployed in a container framework such as Kubernetes, the Hexa-Orchestrator API is not normally exposed for external
ingress and secrets, such as the `hexaclient` secret would be configured as a K8S secret.

Hexa-Orchestrator stores integration information in a file designated by the environment variable `ORCHESTRATOR_CONFIG_FILE`. In
production, this file needs to handled as a secret(Eg. K8S Secret file).  Note that typically integrations to platforms such as Azure, Google, AWS use
credential tokens which are stored in these integration data structures. The credentials used in integrations should be ones issued for Orchestrator's exclusive use. 
These secrets should never be re-used by any other service.

## PostgreSQL

Keycloak uses PostgreSQL to store realm data. Hexa Orchestrator no longer uses PostgreSQL.

