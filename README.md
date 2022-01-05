# Hexa Policy Orchestrator

Hexa is the open-source, standards-based Policy Orchestration software for multi-cloud and hybrid businesses. 

The project contains - 
* Policy Administrator web application
* Policy Orchestrator server
* Demo application

## Development

Ensure the test suite passes.

```bash
go clean -testcache && go test -p 1 ./.../
```

### Run the applications

Run the Hexa Policy Admin web application locally.

```bash
go run cmd/admin/admin.go
```

Run the Hexa Policy Orchestrator server locally.

```bash
go run cmd/orchestrator/orchestrator.go
```

Run the demo web application locally.

```bash
go run cmd/demo/demo.go
```

### Run the applications with docker.

Build the image with pack.

```bash
pack build hexa --builder heroku/buildpacks:20
```

Create `.env` and set the `ORCHESTRATOR_KEY` and `HEXA_DEMO_URL` value.

```bash
ORCHESTRATOR_KEY="<hawk key>"
HEXA_DEMO_URL=http://hexa-demo:8886
```

*Note - use `http://hexa-demo:8886` as above*

Run the Hexa Policy Orchestrator, Admin server, and Demo application via docker compose.

```bash
docker-compose up
```

*Note - include `--platform linux/amd64 ` for M1 processors.*

## Public Cloud setup

Just some quick notes below for Google's Cloud Platform.

Authenticate.

```bash
gcloud auth login
```

Create or use an existing project.

Edit `.env_<environment>.sh` and update the below environment variables.

Source the env file `% source .env_<environment>.sh`.

```bash
gcloud projects create ${GCP_PROJECT_ID} \
  --name ${GCP_PROJECT_NAME} \
  --folder=${GCP_PROJECT_FOLDER} \
  --quiet
```

List the newly create project.

```bash
gcloud projects list
```

Update the `${GCP_PROJECT_NUMBER}` and source the env file again
`% source .env_<environment>.sh` and then set the project.

```bash
gcloud config set project ${GCP_PROJECT_ID}
```

Ensure billing is enabled.

```bash
gcloud services enable cloudbilling.googleapis.com
gcloud alpha billing projects link ${GCP_PROJECT_ID} --billing-account ${GCP_BILLING_ACCOUNT}
```

Other supporting apis

```bash
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
```

Build via Cloud Build

```bash
gcloud builds submit --pack image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1,builder=heroku/buildpacks:20
```

Deploy via Cloud Run

```bash
gcloud run deploy ${GCP_PROJECT_NAME}-policy-admin --command="admin" --region=${GCP_PROJECT_REGION} --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1
gcloud run deploy ${GCP_PROJECT_NAME}-demo --command="demo" --region=${GCP_PROJECT_REGION} --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1
gcloud run deploy ${GCP_PROJECT_NAME}-policy-orchestrator --command="orchestrator" --region=${GCP_PROJECT_REGION} --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1
```

*Note - both web and api server are pack within the same docker image.*

Open policy agent support the demo application.

Build with configuration via docker.

```bash
source .env_development.sh
cd opa-server
docker pull openpolicyagent/opa:latest
docker build --build-arg GCP_PROJECT_ID=${GCP_PROJECT_ID} -t ${GCP_PROJECT_NAME}-opa-server:latest .
docker tag ${GCP_PROJECT_NAME}-opa-server:latest gcr.io/${GCP_PROJECT_ID}/hexa-opa-server:latest
docker push gcr.io/${GCP_PROJECT_ID}/hexa-opa-server:latest
```

Deploy via Cloud Run

```bash
gcloud beta run deploy ${GCP_PROJECT_NAME}-opa-server --region=${GCP_PROJECT_REGION} --image=gcr.io/${GCP_PROJECT_ID}/opa-server:latest \
  --port=8887 --args='--server,--addr,0.0.0.0:8887,--config-file,/config.yaml'
```

We'll need to update the `hexa-opa-server` application environment variable with the `HEXA_DEMO_URL`.

For example `https://<hexa-demo-url>`.

We'll also need to update the `hexa-demo` application environment variable with the `OPA_SERVER_URL`.

For example `https://<opa-server-url>/v1/data/authz/allow`.
