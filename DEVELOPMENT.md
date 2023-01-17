# Development

**NOTE:**

The development tasks herein are also made available via (optional) bash "CLI"
utilities within the repository. Once the initial bootstrapping setup has been
run (via `./bin/pkg.d/setup.sh`), these tasks and more may be executed via the
`dev` and `pkg` CLIs.

For example, try running the following from anywhere within the repository
(assuming the prerequisite "setup" has been run).

```bash
$ dev version
$ dev --help

$ pkg version
$ pkg --help
```

## Task: Bootstrap

> This task may optionally be completed (see **NOTE** above) via:
>
> 1. `pkg setup`
> 2. `dev setup --target=opa`

Clone or download this codebase from GitHub to your local machine and install
the following prerequisites:

* [Go 1.18](https://go.dev)
* [Pack](https://buildpacks.io)
* [Docker Desktop](https://www.docker.com/products/docker-desktop)
* [Open Policy Agent](https://www.openpolicyagent.org) (OPA)
* [PostgreSQL](https://www.postgresql.org/)
* [golang-migrate](https://github.com/golang-migrate/migrate)

## Task: Set up a "test" DB

> This task may optionally be completed (see **NOTE** above) via:
>
> - `dev setup --target=db`

Create a test database in PostgreSQL:

```bash
createuser orchestrator
createdb orchestrator_development --owner orchestrator
psql --command="alter user orchestrator with password 'orchestrator'"
psql --command="grant all privileges on database orchestrator_test to orchestrator"
```

Run the DB migrations:

```bash
migrate -verbose -path ./databases/orchestrator -database "postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable" up
```

## Task: Run the test suite

> This task may optionally be completed (see **NOTE** above) via:
>
> - `dev test`
> - `dev test --clean`

Before making your contributions, ensure the test suite passes:

```bash
go test  ./.../
```

Use the following command to clean up your test cache when needed.

```bash
go clean -testcache
```

## Task: Run the Hexa Applications

> This task may optionally be completed (see **NOTE** above) via:
>
> - `dev serve`

### Steps

- Load the environment variables defined in the `.env_development` file:

  ```bash
  source .env_development
  ```

- Run the Hexa Policy Admin web application.

  ```bash
  go run cmd/admin/admin.go
  ```

- Run the Hexa Policy Orchestrator server.

  ```bash
  go run cmd/orchestrator/orchestrator.go
  ```

- Run the demo web application locally.

  ```bash
  go run cmd/demo/demo.go
  ```

- Run the demo config application locally.

  ```bash
  go run cmd/democonfig/democonfig.go
  ```

- Run the open policy agent server locally.

  ```bash
  opa run --server --addr :8887 -c deployments/opa-server/config/config.yaml
  ```

## CodeQL

GitHub CodeQL is used in the Hexa CI pipeline for vulnerability scanning.
CodeQL can also be installed and run locally on a developer workstation.

To run locally:

- Install via [Homebrew Formulae](https://formulae.brew.sh)

  ```bash
  brew install codeql
  ```

- Install the CodeQL "packs" for Go analysis. The packs and an installation
  script are located in the [CodeQL repository](https://github.com/github/codeql)
  and must be cloned locally.

  ```bash
  cd $HOME/workspace
  git clone https://github.com/github/codeql
  ./codeql/go/scripts/install-deps.sh
  ```

- Create a local database.

  ```bash
  cd $HOME/workspace/policy-orchestrator/
  CODEQL_EXTRACTOR_GO_BUILD_TRACING=on codeql database create .codeql --language=go
  ```

- Analyze the results.

  ```bash
  codeql database analyze .codeql --off-heap-ram=0 --format=csv --output=codeql-results.csv ../codeql/go/ql/src/codeql-suites/go-security-and-quality.qls
  ```
