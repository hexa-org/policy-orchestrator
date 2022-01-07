![hexa-logo](docs/img/hexa-logo.svg) 

# Hexa Policy Orchestrator

[![Build results](https://github.com/hexa-org/almostopen/workflows/build/badge.svg)](https://github.com/hexa-org/almostopen/actions)

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
git clone git@github.com:hexa-org/almostopen.git
```

Build a Hexa image with Pack. The newly created image will contain the policy administrator web application,
policy orchestrator server, and demo application.

```bash
pack build hexa --builder heroku/buildpacks:20
```

Run all three applications with docker compose.

```bash
docker-compose up
```

## Maintainers

* [Hexa project maintainers](maintainers.md)

## Roadmap

* [Current roadmap](roadmap.md)

## Contributing

* [Information for contributors](contributing.md)
* [Additional development information](development.md) 
* [Our of conduct statement](code_of_conduct.md)
