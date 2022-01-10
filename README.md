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

We'll be using postgresql and need to execute the below shell scripts from docker-compose.

```bash
chmod 775 ./databases/docker_support/initdb.d/create-databases.sh
chmod 775 ./databases/docker_support/migrate-databases.sh
```

Run all three applications with docker compose.

```bash
docker-compose up
```

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

## Maintainers

* [Hexa project maintainers](MAINTAINERS.md)

## Roadmap

* [Current roadmap](ROADMAP.md)

## Contributing

[Join the Hexa community](https://hexaorchestration.org/preview/#join) to stay up-to-date with the project and contribute.

* [Additional development information](DEVELOPMENT.md) 
* [Our of conduct statement](CODE_OF_CONDUCT.md)
