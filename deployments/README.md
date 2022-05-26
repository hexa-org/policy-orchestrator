# Hexa applications

_Pardon the dust - the below is work in progress_

We are using [fresh cloud](https://github.com/initialcapacity/freshcloud) to help deploy the hexa suite of applications.

## Google Cloud

## Install Google Cloud SDK

Install the Google Cloud SDK CLI following [these instructions](https://cloud.google.com/sdk/docs/install) or with
[Homebrew](https://formulae.brew.sh/cask/google-cloud-sdk).

## Google Cloud Project Setup

Log in to Google Cloud.

```bash
gcloud auth login
```

Create a service account.

```bash
freshctl clusters gcp create-service-account
```

Create a `.env_apps_google_cloud.sh` file similar to the below.

```bash
export GCP_PROJECT_ID=your_project_id
export GCP_ZONE=your_zone
export GCP_CLUSTER_NAME=your_app_cluster_name
export GCP_SERVICE_ACCOUNT_JSON=.freshcloud/your_service_account.json

export REGISTRY_DOMAIN='your_registry_domain'
export REGISTRY_PASSWORD='your_password'
export REGISTRY_CLUSTER_NAME='your_infra_cluster_name'

export DOMAIN='google.your_domain'
export EMAIL_ADDRESS=your_email
export PASSWORD=your_password

export APP_NAME='hexa-demo'
export APP_IMAGE_NAME='hexa'
export APP_CONFIGURATION_PATH='deployments/google-cloud/hexa-demo.yaml'
export APP_PIPELINE_PATH='deployments/google-cloud/hexa-demo-pipeline.yaml'
export APP_PIPELINE_CONFIGURATION_PATH='deployments/google-cloud/hexa-demo-pipeline-configuration.yaml'
```

Next, source environment the file.

```bash
source .env_apps_google_cloud.sh
```

Create an application cluster.

```bash
freshctl clusters gcp create
```

Create an oauth client.

```bash
gcloud alpha iap oauth-clients create $(gcloud alpha iap oauth-brands list | grep name | sed -e "s/^name: //") --display_name=hexa-demo
```

Create a kubernetes secret for your oauth client using the newly generated `client_id` and `client_secret`.

```bash
kubectl create secret generic hexa-demo-secret \
  --from-literal=client_id=your_client_id \
  --from-literal=client_secret=your_client_secret \
  --namespace='hexa-demo'
```

We'll need a static IP address for the cluster.

```bash
gcloud compute addresses create hexa-demo-static-ip --global --ip-version IPV4
```

Configure the cluster.

```bash
freshctl clusters gcp configure
```

Deploy the demo application.

```bash
freshctl applications deploy  
```

That's a wrap for now.
