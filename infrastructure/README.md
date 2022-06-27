# Hexa infrastructure

Hexa uses fresh cloud to deploy the hexa-demo infrastructure. You could find
out more about [fresh cloud](https://github.com/initialcapacity/freshcloud) on the GitHub page. 

### Current demo topology

| Platform            |               |                                         |
|---------------------|---------------|-----------------------------------------|
| [Google GKE]        |               |                                         |
|                     | Use case      | cloud based hexa-admin                  |
|                     | Apps deployed | hexa-admin and hexa-orchestrator        |
|                     | Deployed      | within a kubernetes clusters            |
|                     | Authorization | not available                           |
|                     | Automation    | concourse pipeline deployed             |
| [Google Cloud IAP]  |               |                                         |
|                     | Use case      | Google's Identity Aware Proxy           |
|                     | Apps deployed | hexa-demo, hexa-demo-config, opa server |
|                     | Deployed      | within a kubernetes clusters            |
|                     | Authorization | reads IAP forward headers               |
|                     | Automation    | concourse pipeline deployed             |
| [Azure AKS]         |               |                                         |
|                     | Use case      | fine-grained policy with OPA            |
|                     | Apps deployed | hexa-demo, hexa-demo-config, opa server |
|                     | Deployed      | within a kubernetes clusters            |
|                     | Authorization | not available                           |
|                     | Automation    | (pending) concourse pipeline deployed   |
| [Azure Authn/Authz] |               |                                         |
|                     | Use case      | coarse grained policy with authn/authz  |
|                     | Apps deployed | hexa-demo, hexa-demo-config, opa server |
|                     | Deployed      | via app services                        |
|                     | Authorization | reads Azure forward headers             |
|                     | Automation    | (pending) manual steps via readme       |
| [AWS Cognito]       |               |                                         |
|                     | Use case      | coarse grained policy with cognito      |
|                     | Apps deployed | hexa-demo, hexa-demo-config, opa server |
|                     | Deployed      | via a kubernetes cluster                |
|                     | Authorization | reads Cognito forward headers           |
|                     | Automation    | concourse pipeline deployed             |
| Local               |               |                                         |
|                     | Use case      | fine-grained policy with OPA            |
|                     | Apps deployed | hexa-demo, hexa-demo-config, opa server |
|                     | Deployed      | via docker compose                      |
|                     | Authorization | not available                           |
|                     | Automation    | via pack locally                        |

The below notes summarize the steps used to create the infrastructure management cluster.

## Install Google Cloud SDK

Install the Google Cloud SDK CLI following [these instructions](https://cloud.google.com/sdk/docs/install) or with
[Homebrew](https://formulae.brew.sh/cask/google-cloud-sdk).

## Management cluster

Create a `.env_infra.sh` file similar to the below.

```bash
export GCP_PROJECT_ID=your_project_id
export GCP_ZONE=your_zone
export GCP_CLUSTER_NAME=your_cluster_name

export DOMAIN=your_domain
export EMAIL_ADDRESS=your_email
export PASSWORD=your_password
```

Next, source environment the file.

```bash
source .env_infra.sh
```

Log in to Google Cloud.

```bash
gcloud auth login
```

Configure your google cloud project.

```bash
gcloud config set project ${GCP_PROJECT_ID}
```

Ensure the project was set correctly.

```bash
gcloud projects describe ${GCP_PROJECT_ID}
```

Download fresh cloud resources locally.

```bash
freshctl resources copy
```

Update and source your `.env_infra.sh` with the resources directory.

```bash
export FRESH_RESOURCES=your_local_resources_directory
```

Local resources are now available for customization.

Create a management cluster.

_Note adding the `--execute` flag will execute the command._

```bash
freshctl clusters gcp enable-services
freshctl clusters gcp create
freshctl services add contour
```

Create a DNS entry for your load balancer. Re-run the below command to show your ip address as needed.

```bash
kubectl describe svc ingress-contour-envoy --namespace projectcontour | grep Ingress | awk '{print $3}'
```

Continue with management services.

```bash
freshctl services add cert-manager
freshctl services add harbor
freshctl services add concourse
freshctl services add kpack
```

Confirm the management cluster services are deployed.

* Harbor https://registry.your_domain
* Concourse https://ci.your_domain

That's a wrap for now.

[Google GKE]:https://cloud.google.com/kubernetes-engine
[Google Cloud IAP]:https://cloud.google.com/iap
[Azure AKS]:https://azure.microsoft.com/en-us/services/kubernetes-service/#overview
[Azure Authn/Authz]:https://docs.microsoft.com/en-us/azure/app-service/overview-authentication-authorization
[AWS Cognito]:https://aws.amazon.com/cognito/