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

### Hexa-Orchestrator

Hexa Orchestrator is the internal API gateway through which policy provisioning and retrieval takes place. Requests from
Hexa-Admin and other container clients are secured using [Hawk](https://github.com/mozilla/hawk) which implements a form
of [digest message authentication](https://github.com/mozilla/hawk/blob/main/API.md).

_Note_ The credentials used within this repository are for testing. Please create new credentials for your environment.

In the current implementation HAWK verifies that the request URI used by the client matches the configured request URI.
You will need to ensure Hexa-Admin and Hexa-Orchestrator use the same domain and port numbers or requests will fail.

When deployed in a container framework such as Kubernetes, the Hexa-Orcha API is not normally exposed for external
ingress.

Eventually the orchestrator shall be expanded to support common API authentication methods such as JWT tokens, PATs, and
sender constrained tokens such as [OAuth DPOP](https://datatracker.ietf.org/doc/draft-ietf-oauth-dpop/).

## PostgreSQL

Hexa currently uses PostgreSQL to store provider service account credentials. Please ensure your PostgreSQL instance and
client connection is secure. More information can be found below -

* [encryption-options](https://www.postgresql.org/docs/8.1/encryption-options.html)
* [ssl-tcp](https://www.postgresql.org/docs/current/ssl-tcp.html)

The docker-compose configuration does not normally expose the database outside of docker. If you get a network
connection refused message in Hexa Orchestrator (e.g. beause Hexa Orchestrator is not running inside Docker), then port
5432 will need to be exposed by the postgresql container by revising your docker-compose.yml.
