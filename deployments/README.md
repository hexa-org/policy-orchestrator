# Hexa applications

Hexa uses fresh cloud to deploy the hexa-demo applications. You could find
out more about [fresh cloud](https://github.com/initialcapacity/freshcloud) on the GitHub page.

The below notes summarize the steps used to deploy application pipelines.

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

## Hexa-demo

Create a `.env_google_cloud_demo.sh` file similar to the below.

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
source .env_google_cloud_demo.sh
```

Create an application cluster.

```bash
freshctl clusters gcp create
```

Create an oauth client.

```bash
gcloud alpha iap oauth-clients create $(gcloud alpha iap oauth-brands list | grep name | sed -e "s/^name: //") --display_name=${APP_NAME}
```

Create a kubernetes secret for your oauth client using the newly generated `client_id` and `client_secret`.

```bash
kubectl create namespace ${APP_NAME}
kubectl create secret generic ${APP_NAME}-secret \
  --from-literal=client_id=your_client_id \
  --from-literal=client_secret=your_client_secret \
  --namespace=${APP_NAME}
```

Create a static IP address for the cluster.

```bash
gcloud compute addresses create ${APP_NAME}-static-ip --global --ip-version IPV4
```

Configure the cluster.

```bash
freshctl clusters gcp configure
```

Deploy the demo application.

```bash
freshctl applications deploy
```

## Hexa-admin

Follow the above steps for the hexa-admin application - although, select a different name for your environment file
`.env_google_cloud_admin.sh` file similar to the below.

Install a postgresql database for the orchestrator.

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm install ${APP_NAME}-db bitnami/postgresql \
    --namespace ${APP_NAME} \
    --set persistence.existingClaim=${APP_NAME}-pvc \
    --set primary.resources.requests.cpu=0 \
    --set volumePermissions.enabled=true
```

Create a database for the orchestrator.

```bash
export POSTGRES_PASSWORD=$(kubectl get secret --namespace ${APP_NAME} ${APP_NAME}-db-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)
```

```bash
kubectl run ${APP_NAME}-db-postgresql-client --rm --tty -i --restart='Never' --namespace ${APP_NAME} --image docker.io/bitnami/postgresql:14.3.0-debian-10-r17 \
  --env="PGPASSWORD=$POSTGRES_PASSWORD" \
  --command -- psql --host ${APP_NAME}-db-postgresql -U postgres -d postgres -p 5432   
```

```sql
create database orchestrator_development;
create user orchestrator with password 'orchestrator';
grant all privileges on database orchestrator_development to orchestrator;
```

Run the database schema migration scripts.

```bash
kubectl run ${APP_NAME}-db-postgresql-migrate -it --namespace ${APP_NAME} --image migrate/migrate --command sh 
kubectl cp databases/orchestrator ${APP_NAME}/${APP_NAME}-db-postgresql-migrate:/home/orchestrator --namespace ${APP_NAME} 
kubectl exec ${APP_NAME}-db-postgresql-migrate -it --namespace ${APP_NAME} sh
/  migrate -verbose -path /home/orchestrator -database postgres://orchestrator:orchestrator@hexa-admin-db-postgresql:5432/orchestrator_development?sslmode=disable up
```

That's a wrap for now.
