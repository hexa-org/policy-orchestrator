# Deploy Hexa Policy Orchestration to GCP

## Google Cloud Project Setup

Log in to Google Cloud.

```bash
gcloud auth login
```

Create .env_<environment>.sh file to store environment variables.

```bash
export GCP_PROJECT_NAME=<gcp project name>
export GCP_PROJECT_FOLDER=<gcp project folder>
export GCP_PROJECT_ID=<gcp project id>
export GCP_PROJECT_REGION=<gcp region>
export GCP_BILLING_ACCOUNT=<billing account id>
```

Source the env file

```bash
source .env_<environment>.sh
```

Create a GCP project.

```bash
gcloud projects create ${GCP_PROJECT_ID} \
  --name ${GCP_PROJECT_NAME} \
  --folder=${GCP_PROJECT_FOLDER} \
  --quiet
```

View the newly created project.

```bash
gcloud projects list
```

Configure the Google Cloud CLI to use your new project.

```bash
gcloud config set project ${GCP_PROJECT_ID}
```

Ensure billing is enabled.

```bash
gcloud services enable cloudbilling.googleapis.com
gcloud alpha billing projects link ${GCP_PROJECT_ID} --billing-account ${GCP_BILLING_ACCOUNT}
```

Enable other supporting APIs.

```bash
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable compute.googleapis.com
gcloud services enable vpcaccess.googleapis.com
```

## Deploy via Cloud Run

Build Image via Cloud Build

```bash
gcloud builds submit --pack image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1,builder=heroku/buildpacks:20
```

Deploy Hexa Policy Admin.

```bash
gcloud run deploy ${GCP_PROJECT_NAME}-policy-admin \
  --command="admin" \
  --region=${GCP_PROJECT_REGION} \
  --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1
```

Deploy Hexa Policy Orchestrator.

```bash
gcloud run deploy ${GCP_PROJECT_NAME}-policy-orchestrator \
  --command="orchestrator" \
  --region=${GCP_PROJECT_REGION} \
  --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1 \
  --ingress internal 
```

Both web and api server are packed within the same docker image.

### Set up networking

The Policy Admin app will be accessible via the internet but the Policy Orchestrator will not be. To achieve this, we
have [restricted ingress](https://cloud.google.com/run/docs/securing/ingress) to the Policy Orchestrator app by
deploying it with the `--ingress internal`.

To allow the Policy Admin app to communicate with the Policy Orchestrator, first create a new VPC network connector

```bash
gcloud compute networks vpc-access connectors create hexa-vpc \
  --network default \
  --region ${GCP_PROJECT_REGION} \
  --range 10.8.0.0/28 
```

Then bind the VPC network connector to the Policy Admin app and configure egress.

```bash
gcloud run services update hexa-policy-admin \
  --region ${GCP_PROJECT_REGION} \
  --vpc-connector hexa-vpc \
  --vpc-egress all-traffic
``` 

### Deploy Demo App and OPA Server

Deploy demo app.

 ```bash
 gcloud run deploy ${GCP_PROJECT_NAME}-demo --command="demo" --region=${GCP_PROJECT_REGION} --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1
 ```

Build OPA Agent with configuration via Docker.

```bash
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

Update the `hexa-opa-server` application environment variable with the `HEXA_DEMO_URL`. For
example `https://<hexa-demo-url>`.

Update the `hexa-demo` application environment variable with the `OPA_SERVER_URL`. For
example `https://<opa-server-url>/v1/data/authz/allow`.
