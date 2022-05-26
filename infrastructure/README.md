# Hexa infrastructure

_Pardon the dust - the below is work in progress_

We are using [fresh cloud](https://github.com/initialcapacity/freshcloud) to help stand up our hexa infrastructure. 

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
freshctl clusters gcp list
freshctl services contour
```

Create a DNS entry for your load balancer. Re-run the below command to show your ip address.

```bash
kubectl describe svc ingress-contour-envoy --namespace projectcontour | grep Ingress | awk '{print $3}'
```

Continue with management services.

```bash
freshctl services cert-manager
freshctl services harbor
freshctl services concourse
freshctl services kpack
```

Confirm the management cluster services are deployed.

* Harbor https://registry.your_domain
* Concourse  https://ci.your_domain


That's a wrap for now.
