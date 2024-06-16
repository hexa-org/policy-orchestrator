# Development

> [!Note]
> This documentation is currently out of date and will be updated shortly.
 
To install the environment, if not already performed, follow the instructions on the [README page](README.md).


## Task: Orchestrator endpoints via Postman
Hexa Orchestrator uses [Hawk Authentication](https://github.com/mozilla/hawk/blob/main/API.md) to allow clients to make authenticated requests.

During Authentication, the client sends a MAC (Message Authentication Code) as part of the Hawk Authorization Header.

The MAC is calculated using HMAC with SHA256 hashing over the `normalized request string`.

The `normalized reqest string` is made up of the HOST, PORT amongst other things.

If you run the Hexa Policy Orchestrator using `docker-compose up` and try to hit an orchestrator endpoint via Postman `localhost:8885`, you will get a `401 Unauthorized`
- This is because Postman calculates the MAC using `localhost:8885`.
- HOST and PORT that Orchestrator users to verify the MAC is specified by the `ORCHESTRATOR_HOSTPORT` environment variable in `docker-compose.yml`

### Steps
To hit the Orchestrator endpoints via Postman:
- Remove the `ORCHESTRATOR_HOSTPORT: hexa-orchestrator:8885` environment variable from the `docker-compose.yml`

NOTE: This will cause problems if you are running the admin app to do things with the orchestrator. 
The admin app uses the orchestrator host and port when sending requests so the key would need to be added back at this time.

See [Hawk authorization failing with Postman](https://github.com/hexa-org/policy-orchestrator/issues/261) for details.

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
