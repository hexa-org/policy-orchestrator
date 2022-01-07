# Development 

Clone or download the codebase from GitHub to your local machine and install the following prerequisites.

* [Go 1.17](https://go.dev)
* [Pack](https://buildpacks.io)
* [Docker Desktop](https://www.docker.com/products/docker-desktop)
* [Open policy agent](https://www.openpolicyagent.org)

```bash
cd ~/workspace
git clone git@github.com:hexa-org/almostopen.git
```

Install via [Homebrew Formulae](https://formulae.brew.sh)

```bash
brew install go pack opa docker docker-compose
```

## Run the tests

Ensure the test suite passes.

```bash
go test -p 1 ./.../
```

Use the following command to clean up your test cache when needed.

```bash
go clean -testcache
```

### Run the applications

Source the `.env_development` file.

```bash
source .env_development
```

Run the Hexa Policy Admin web application.

```bash
go run cmd/admin/admin.go
```

Run the Hexa Policy Orchestrator server.

```bash
go run cmd/orchestrator/orchestrator.go
```

Run the demo web application locally.

```bash
go run cmd/demo/demo.go
```

Run the open policy agent server locally.

```bash
opa run --server --addr :8887 -c deployments/opa-server/config/config.yaml
```
