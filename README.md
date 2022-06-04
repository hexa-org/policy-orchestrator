![hexa-logo](docs/img/hexa-logo.svg)

# Hexa Policy Orchestrator

[![Build results](https://github.com/hexa-org/policy-orchestrator/workflows/build/badge.svg)](https://github.com/hexa-org/policy-orchestrator/actions)
[![Go Report Card](https://goreportcard.com/badge/hexa-org/policy-orchestrator)](https://goreportcard.com/report/hexa-org/policy-orchestrator)
[![codecov](https://codecov.io/gh/hexa-org/policy-orchestrator/branch/main/graph/badge.svg)](https://codecov.io/gh/hexa-org/policy-orchestrator)
[![CodeQL](https://github.com/hexa-org/policy-orchestrator/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/hexa-org/policy-orchestrator/actions/workflows/codeql-analysis.yml)

Hexa is the open-source, standards-based policy orchestration software for multi-cloud and hybrid businesses.

The Hexa project contains three applications.
* Policy Administrator web application
* Policy Orchestrator server with IDQL translations
* Demo application

Hexa Policy Orchestration (Hexa) and Identity Query Language (IDQL) were purpose-built to solve the proliferation of
policy orchestration problems caused by today’s hybrid cloud and multi-cloud world. Together, Hexa and IDQL enable you
to manage all of your policies consistently across clouds and vendors so you can unify access policy management.

## Getting Started

Clone or download the codebase from GitHub to your local machine and install the following prerequisites.

* [Go 1.17](https://go.dev)
* [Pack](https://buildpacks.io)
* [Docker Desktop](https://www.docker.com/products/docker-desktop)

```bash
cd /home/user/workspace/
git clone git@github.com:hexa-org/policy-orchestrator.git
```

### Build

Build a Hexa image with Pack. The newly created image will contain the policy
administrator web application, policy orchestrator server, and demo application.

```bash
pack build hexa --builder heroku/buildpacks:20
```

We'll be using postgresql and need to execute the below shell scripts from docker-compose.

```bash
chmod 775 ./databases/docker_support/initdb.d/create-databases.sh
chmod 775 ./databases/docker_support/migrate-databases.sh
```

### Run

Run all the applications with docker compose.

```bash
docker-compose up
```

Docker runs the applications described below.

**hexa-orchestrator** runs on [localhost:8885](http://localhost:8885/health). The main application
that manages IDQL policy across various platforms and communicates with the various platform interfaces;
converting IDQL policy to and from the respective platform types.

**hexa-admin** runs on [localhost:8884](http://localhost:8884/). An example application
demonstrating the latest interactions with the policy orchestrator.

**hexa-demo** runs on [localhost:8886](http://localhost:8886/). A demo application used to highlight
enforcing both coarse and fine-grained policy. The application integrates with platform
authentication/ authorization proxies, [Google IAP](https://cloud.google.com/iap) for example,
for coarse grained access and the [Open Policy Agent (OPA)](https://www.openpolicyagent.org/)
for fine-grained policy access.

**OPA server** runs on [localhost:8887](http://localhost:8887/). The Open Policy Agent (OPA) server used to 
demonstrate fine-grained policy management. IDQL policy is represented as data and interpreted by
the rego expression language.

**hexa-demo-config** runs on [localhost:8889](http://localhost:8889/health). The bundle HTTP server from which the
OPA server can download the bundles of policy and data from. See [OPA bundles][opa-bundles] for more info.

### Example workflow

Using the hexa-admin application available via docker-compose, upload an OPA integration
configuration file. The file describes the location of the IDQL policy.

```json
{
  "bundle_url":"http://hexa-demo-config:8889/bundles/bundle.tar.gz"
}
```

Once configured, IDQL policy for the hexa-demo application can be modified on
the Applications page. The hexa-admin communicates the changes to the
hexa-orchestrator or **policy management point** which then updates the hexa-demo-config bundle server -
making the updated policy available to the OPA server.

OPA or **policy decision point** periodically reads config from the hexa-demo-config bundle
server and allows or denies access requests based on the IDQL policy.
Decision enforcement is handled within the hexa-demo application or **policy enforcement point**.

![Hexa Demo Architecture](docs/img/Hexa-Demo-Architecture.png "hexa demo architecture")

### Cleanup

Cleaning up. Remove all docker containers and volumes.

```bash
docker rm -f $(docker ps -a -q)
docker volume rm -f $(docker volume ls -q)
docker system prune -a -f
```

Remove the local postgres database files.

```bash
rm -rf .postgres
```

## Cloud Native Computing Foundation

Hexa uses the below Cloud Native Computing Foundation ([CNCF](https://www.cncf.io/)) projects

* [Contour](https://projectcontour.io/)
* [Harbor](https://goharbor.io/)
* [Helm](https://helm.sh/)
* [Kubernetes](https://kubernetes.io/)
* [Open Policy Agent](https://www.openpolicyagent.org/)
* [Pack](https://buildpacks.io/)
* [Prometheus](https://prometheus.io/)

The current demo deployment infrastructure can be found at this [link](infrastructure/README.md).

[opa-bundles]: https://www.openpolicyagent.org/docs/latest/management-bundles/