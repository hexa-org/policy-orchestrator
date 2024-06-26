# Table of Contents

- [Hexa Policy Orchestrator](#hexa-policy-orchestrator)
    * [Getting Started](#getting-started)
        + [Build the Hexa image](#task-build-the-hexa-orchestrator-image)
        + [Run the Policy Orchestrator](#task-run-the-policy-orchestrator)
    * [Application descriptions](#application-descriptions)
        + [Example workflow](#example-workflow)
    * [Getting involved](#getting-involved)

![hexa-logo](docs/hexa-logo.svg)

# Hexa Policy Orchestrator

[![Build results](https://github.com/hexa-org/policy-orchestrator/workflows/build/badge.svg)](https://github.com/hexa-org/policy-orchestrator/actions)
[![Go Report Card](https://goreportcard.com/badge/hexa-org/policy-orchestrator)](https://goreportcard.com/report/hexa-org/policy-orchestrator)
[![codecov](https://codecov.io/gh/hexa-org/policy-orchestrator/branch/main/graph/badge.svg)](https://codecov.io/gh/hexa-org/policy-orchestrator)
[![CodeQL](https://github.com/hexa-org/policy-orchestrator/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/hexa-org/policy-orchestrator/actions/workflows/codeql-analysis.yml)

Hexa Policy Orchestrator enables you to manage all of your policies consistently across software providers
so that you can unify access policy management. The below diagram describes the current provider architecture.

![Hexa Provider Architecture](docs/hexa-provider-architecture.svg "hexa provider architecture")

## Getting Started

The Hexa project contains two applications, and demonstrates use of applications
from [Policy-Opa](https://github.com/hexa-org/policy-opa)

- Policy Orchestrator with policy translations
- Demo Policy Administrator
- Demo web application (Policy-OPA)
- Hexa OPA Server and Bundle Server (Policy-OPA)

To get started with running these, clone or download the codebase from GitHub to your local machine:

```bash
cd $HOME/workspace # or similar
git clone git@github.com:hexa-org/policy-orchestrator.git
```

### Prerequisites

Install the following dependencies.

- [Go 1.22](https://go.dev) - Needed to compile and install
- [Docker Desktop](https://www.docker.com/products/docker-desktop) - Needed to run docker-compose configuration

### Task: Build the Hexa Orchestrator image

Build the Hexa Orchestrator Docker containers...

```bash
cd demo
sh ./build.sh
```

### Task: Run the Policy Orchestrator

Run all the applications with Docker Compose from within the `demo` directory

```bash
docker-compose up
```

## Application Descriptions

Docker runs the following applications:

- **hexa-orchestrator**

  Runs on [localhost:8885](http://localhost:8885/health). The main application service API
  that manages IDQL policy across various platforms and communicates with the
  various platform interfaces, converting IDQL policy to and from the respective
  platform types using the [Hexa Policy-Mapper](https://github.com/hexa-org/policy-mapper) SDK. This service uses TLS
  and is protected and uses JWT authorization tokens
  (see [RFC7519](https://datatracker.ietf.org/doc/html/rfc7519)) for authenticating access.

- **hexa-admin-ui**

  Runs on [localhost:8884](http://localhost:8884/). An example web application user-interface
  demonstrating the latest interactions with the policy orchestrator. Uses `hexa-orchestrator` to retrieve and provision
  policy.
  <p/>
  The Hexa Admin UI service uses the OAuth2 Client Credentials Grant Flow (see [RFC6749 Section 4.4](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4)) to authenticate to a token server (e.g. Keycloak) and obtain tokens for accessing
  the Hexa-Orchestrator API service using JWT tokens.

- **hexa-industry-demo-app**

  Runs on [localhost:8886](http://localhost:8886/). A demo web application used to
  highlight enforcing of both coarse and fine-grained policy. The application
  integrates with platform authentication/authorization proxies,
  [Google IAP](https://cloud.google.com/iap) for example, for coarse-grained
  access and the [Open Policy Agent (OPA)](https://www.openpolicyagent.org/)
  for fine-grained policy access. In the docker-compose configuration, the server uses the Hexa-OPA-Server

- **hexa-opa-agent**

  Runs on [localhost:8887](http://localhost:8887/). HexaOPA is
  an [IDQL extended](https://github.com/hexa-org/policy-opa) Open Policy Agent ([OPA](https://openpolicyagent.org))
  server used as an IDQL Policy Decision service. In this configuration, The `hexa-opa-agent` server retrieves policies
  from
  the `hexa-opaBundle-server` instance which is an OPA provisioning service.

- **hexa-opaBundle-server**

  Runs on [localhost:8889](http://localhost:8889/health). An OPA HTTP Bundle server from which the OPA server can
  download policy bundles configured
  by Hexa-Orchestrator. See [OPA bundles][opa-bundles] for more info.

- **keycloak** and **postgres**

  These two servers are used to provide an OAuth2 demonstration and testing environment for Hexa
  services. [Keycloak](https://www.keycloak.org) is pre-configured with a demonstration security "realm" called
  [Hexa-Orchestrator-Realm](http://localhost:8080/admin/master/console/#/Hexa-Orchestrator-Realm). This realm contains
  role definitions and Client credential enabling Hexa Admin UI service to access the Hexa Orchestrator service. By default,
  a bootstrap admin user id and password are configured (see .env_development). It is strongly recommended that these be
  changed.

## Configuration

### TLS Support
In the current release, except for the Hexa Admin UI, and Hexa Industries Demo application, all Hexa services now auto
configure with TLS support. In general if HEXA_SERVER_KEY_PATH is not specified for a server, the startup code will do the 
following:
1. Locks the directory to avoid conflict with other server instances
2. Look for an existing CA key pair
3. If not found, generates a new key pair
4. Generates a server TLS key pair using the DNS names specified. Note: in docker, it is best to give each server instance it's own key file name and DNS name.
5. Releases the directory lock

> [!Note]
> The lock procedure is done to ensure only 1 set of keys is generated so that all servers will share the same self-signed key root.

When a second server starts and obtains a lock, it will use the CA key pair generated by the first server to generate its own server
key pair.

The intent of this procedure to use TLS whenever possible and to make demonstrations and development easy to set up and 
configure. For production use, use externally generated keys from appropriate certificate authorities and assign the key
files to paths using the appropriate environment variable. 

> [!Tip]
> When installing in a cloud service provider that will be doing TLS termination, set the value of `HEXA_TLS_ENABLED` and
> `HEXA_AUTO_SELFSIGN` to false.

### OAuth2 Support

The Hexa Administration Server currently supports unauthenticated access only and should be installed in a protected location. It
will be upgraded shortly to support Open Identify (OIDC) based authentication. 

The Hexa Administration server now makes provisioning calls to Hexa Orchestrator over secured TLS communications using the OAuth2 Client
Credentials Grant ([RFC6749 Section 4.4](https://www.rfc-editor.org/rfc/rfc6749.html#section-4.4)). In the docker-compose file, a demonstration configuration is set-up using [KeyCloak](https://keycloak.org) 
to authenticate client requests and issue tokens for accessing the Hexa Orchestration Service.

> [!Important]
> To facilitate easy setup for demonstration, Keycloak is preconfigured with a realm (Hexa_Orchestrator_Realm), and a client credential. The root
> password to Keylcloak and the `hexaclient` client secret should be changed!  See Keycloak [documentation](https://www.keycloak.org/guides#server).

### Environment Variables

| Name                                                                              | Default         | Description                                                                             |
|-----------------------------------------------------------------------------------|-----------------|-----------------------------------------------------------------------------------------|
| HEXA_TLS_ENABLED                                                                  | false           | Enable TLS if supported (default: False)                                                |
| HEXA_CERT_DIRECTORY                                                               | $HOME/.certs    | Directory where PEM files are stored and generated                                      |
| HEXA_CA_KEYFILE                                                                   | ca-key.pem      | Private key PEM file used to generate self-signed TLS certs                             |
| HEXA_CA_CERT                                                                      | ca-cert.pem     | Certificate Authority Public Key PEM file. Used to validate server certificates         |
| HEXA_SERVER_KEY_PATH                                                              | server-key.pem  | Server private key file used to establish TLS services.                                 |
| HEXA_SERVER_CERT                                                                  | server-cert.pem | TLS Public certificate file for establishing TLS services                               |
| HEXA_SERVER_DNS_NAME                                                              |                 | Comma separated list of DNS names. Used when auto-generating a server TLS certificate   |
| HEXA_AUTO_SELFSIGN                                                                | true            | When set to false, server will not attempt to auto generate self-signed certificaes     |
| HEXA_CERT_ORG, <br/>HEXA_CERT_COUNTRY,<br/>HEXA_CERT_PROV,<br/>HEXA_CERT_LOCALITY |                 | Values to be used when generating TLS certificates                                      |
| HEXA_TOKEN_JWKSURL                                                                | <none>          | When using token based authentication, the URL of the JWKS endpoint                     |
| HEXA_JWT_AUTH_ENABLE                                                              | false           | Enable token based authentication                                                       |
| HEXA_JWT_REALM                                                                    | undefined       | The OAuth2 realm - used in error message response to clients to indicate token issuer   |
| HEXA_JWT_AUDIENCE                                                                 | orchestrator    | The audience value that should be present in received tokens (used by Orchestrator API) |
| HEXA_JWT_SCOPE                                                                    | orchestrator    | Token 'scope' claim value expected (e.g. `orchestrator`)                                |
| HEXA_OAUTH_CLIENT_ID                                                              |                 | The OAuth ClientId value used by Admin UI at the OAuth Token Endpoint                   |
| HEXA_OAUTH_CLIENT_SECRET                                                          |                 | OAuth Client secret used in the client credentials flow                                 |
| HEXA_OAUTH_CLIENT_SCOPE                                                           |                 | If supplied, the scope value to be passed in the Client Credentials Grant request       |
| HEXA_OAUTH_TOKEN_ENDPOINT                                                         |                 | The OAUTH2 Token endpoint URL used to execute the client credentials grant.             |

## Example Workflow

Fine-grained policy management with OPA.

Using the **hexa-admin-ui** application available via `docker-compose`, upload an
OPA integration configuration file. The file describes the location of the IDQL
policy. An example integration configuration file may be found in
[deployments/opa-server/example](demo/deployments/hexaOpaServer/example).

Once configured, IDQL policy for the **hexa-demo** application can be modified
on the [Applications](http://localhost:8884/applications) page. The
**hexa-admin** communicates the changes to the **hexa-orchestrator**, or
"Policy Management Point (PMP)", which then updates the **hexa-demo-config** bundle
server, making the updated policy available to the OPA server.

OPA, the "Policy Decision Point (PDP)", periodically reads config from the
**hexa-demo-config** bundle server and allows or denies access requests based on
the IDQL policy. Decision enforcement is handled within the **hexa-demo**
application or "Policy Enforcement Point (PEP)".

The Hexa Demo architecture may be visualized as follows (note: POSTGRES is no longer used by Hexa-Orchestrator and now
uses a JSON based configuration file specified by `ORCHESTRATOR_CONFIG_FILE`):

![Hexa Demo Architecture](docs/hexa-demo-architecture.svg "hexa demo architecture")

## Getting involved

Take a look at our [product backlog](https://github.com/orgs/hexa-org/projects/1)
where we maintain a fresh supply of good first issues. In addition to
enhancement requests, feel free to post any bugs you may find.

- [Backlog](https://github.com/orgs/hexa-org/projects/1)

Here are a few additional resources for those interested in contributing to the
Hexa project:

- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Development](DEVELOPMENT.md)
- [Security](SECURITY.md)

This [repository](https://github.com/hexa-org/policy-orchestrator) also includes
[documentation](docs/infrastructure/README.md) for the current demo deployment
infrastructure.

[opa-bundles]: https://www.openpolicyagent.org/docs/latest/management-bundles/

<footer>
   <p>The Linux Foundation has registered trademarks and uses trademarks. For a list of trademarks of The Linux Foundation, 
         please see our <a href="https://www.linuxfoundation.org/legal/trademark-usage">Trademark Usage page</a>.
   </p>
</footer>