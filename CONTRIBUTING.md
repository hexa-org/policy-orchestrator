# Hexa Contributor Guide

Welcome to the Hexa community! This guide explains the various ways you can participate in and contribute to the
work of policy orchestration.

### Contributing

[Join the Hexa community](https://hexaorchestration.org/preview/#join) to stay up-to-date with the project and contribute.

* Before you can contribute, please sign the [contributor license agreement](https://docs.linuxfoundation.org/lfx/easycla/contributors)
* Please make sure to read and observe our [our of conduct statement](CODE_OF_CONDUCT.md)

Additional development information can be found [here](DEVELOPMENT.md).

### Project Roles

* **Contributors** are anyone that engages on a regular basis with the Hexa-IDQL project by updating documentation, doing 
code reviews, creating or responding to issues, contributing code, and so on.

* **Maintainers** are the team responsible for overall project governance and project direction. Maintainers have final
approval over PRs and setting priority within the project backlog.
Please [review](https://github.com/hexa-org/policy-orchestrator/blob/main/MAINTAINERS.md) the current list of
maintainers.

### Project Backlog

The project backlog is maintained at this [link](https://github.com/orgs/hexa-org/projects/1/views/4). Here you can see
work in progress as well as the list of items in our backlog. If you want to be assigned to a user story, contact the
maintainers.

### Golang coding conventions

* Intentions should be clear
* We encourage real-time linting environments such as [Goland](https://www.jetbrains.com/go/)
* Functions should _not_ be documented with comments unless describing why; described why, not what
* Ignoring errors should be rare but explicit and marked with an underscore; `db.Exec(` versus `_, _ = db.Exec(`
* Favor passing structs by value not reference; favor immutability
* Avoid locking or sleeping
* Avoid short, ambiguous names; favor clarity
* Support package names should end in `*support`; common pkgs that support your application code, replaceable by an open source framework or library
* We use [Pack](https://buildpacks.io) for building containers
* Ensure everything runs locally via `docker compose`
* Favor a single repository; to move quickly
* Favor replace-ability aka low coupling high cohesion; cmds, pkgs, support pkgs could be replaced with something better over time
* Ensure all code is tested
* Tests should be in their own package `package metricssupport_test` - with no access to private functions
* Test package names should include `*_test`
* Avoid mocking database tests
* Avoid mocking HTTP tests
* Local unit, integration, and acceptance tests should run within a few seconds, not minutes
* Refactor ruthlessly

### Create a new provider

The Hexa project has several providers for integrating with different cloud platforms or systems (Google, Microsoft,
Amazon, etc.). If you would like to extend Hexa and build a new integration - go for it! We will be here to support your
effort along the way. New providers can be added to our existing repo or, we can create a separate repo if that makes
sense.
