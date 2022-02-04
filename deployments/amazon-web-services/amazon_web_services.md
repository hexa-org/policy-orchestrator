# Deploy to Amazon Web Services

## Install the AWS command-line interface

```bash
brew install -q awscli eksctl helm
```

Configure the command-line interface.

```bash
aws configure
 ```

For now, we'll use "us-east-2" for the default region.

Create a `.env_amazon_web_services.sh` file to store your amazon environment variables.

```bash
export AWS_ACCOUNT_ID="<your aws account id>"
export AWS_PROJECT_NAME="<your project name>"
export AWS_REGION="us-east-2"
```

Source the `.env_amazon_web_services.sh` file.

```bash
source .env_amazon_web_services.sh
```

## Build via Pack

Build the hex image.

```bash
pack build hexa --builder heroku/buildpacks:20
```

## Register the image via Elastic Container Registry

Login to ecr via docker.

```bash
PASSWORD=$(aws ecr get-login-password --region $AWS_REGION)
  echo ${PASSWORD} | docker login --username AWS --password-stdin \
    ${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com
```

Create a repository.

```bash
aws ecr create-repository \
  --repository-name ${AWS_PROJECT_NAME}/hexa \
  --image-scanning-configuration scanOnPush=false \
  --image-tag-mutability IMMUTABLE \
  --region ${AWS_REGION}
```

Push the newly built hexa image.

```bash
docker tag hexa ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${AWS_PROJECT_NAME}/hexa:tag1
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${AWS_PROJECT_NAME}/hexa:tag1
```

Build an OPA server with configuration via Docker.

From the `./deployments/opa-server` directory run the below commands.

```bash
docker pull openpolicyagent/opa:latest
docker build --build-arg OPA_PROJECT_ID=${AWS_PROJECT_ID} -t opa-server:latest .
````

Create a repository for the open policy agent server.

```bash
aws ecr create-repository \
  --repository-name ${AWS_PROJECT_NAME}/opa-server \
  --image-scanning-configuration scanOnPush=false \
  --image-tag-mutability IMMUTABLE \
  --region ${AWS_REGION}
```

Push the open policy agent server image.

```bash
docker tag opa-server:latest ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${AWS_PROJECT_NAME}/opa-server:latest
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${AWS_PROJECT_NAME}/opa-server:latest
```

## Deploy via Kubernetes

Deploy the Demo app and the OPA Server to Kubernetes.

Create a new kubernetes cluster.

```bash
envsubst < kubernetes/cluster-config.yaml | eksctl create cluster -f -
```

Write the configuration details as needed.

```bash
aws eks --region ${AWS_REGION} update-kubeconfig --name ${AWS_PROJECT_NAME}
````

Create an IAM Open ID Connect provider.

```bash
eksctl utils associate-iam-oidc-provider --cluster=${AWS_PROJECT_NAME} --region=${AWS_REGION} --approve
```

Create the RBOC roles.

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.4/docs/examples/rbac-role.yaml
```

Create the IAM policy as needed.

```bash
curl -o iam_policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.3.1/docs/install/iam_policy.json
aws iam create-policy \
    --policy-name AWSLoadBalancerControllerIAMPolicy \
    --policy-document file://iam_policy.json
```

Create the service account.

```bash
eksctl create iamserviceaccount \
    --cluster=${AWS_PROJECT_NAME} \
    --namespace=kube-system \
    --name=aws-load-balancer-controller \
    --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT_ID}:policy/AWSLoadBalancerControllerIAMPolicy \
    --override-existing-serviceaccounts \
    --approve
```

Install cert manager.

```bash
kubectl apply \
  --validate=false \
  -f https://github.com/jetstack/cert-manager/releases/download/v1.1.1/cert-manager.yaml
```

Install the ingress controller.

```bash
envsubst < kubernetes/v2_3_1_full.yaml | kubectl apply -f -
```

Deploy demo app objects.

```bash
envsubst < kubernetes/demo/deployment.yaml | kubectl apply -f -
envsubst < kubernetes/demo/service.yaml | kubectl apply -f -
envsubst < kubernetes/demo/ingress.yaml | kubectl apply -f -
````

Deploy demo opa objects.

```bash
envsubst < kubernetes/opa-server/deployment.yaml | kubectl apply -f -
envsubst < kubernetes/opa-server/service.yaml | kubectl apply -f -
````

Cleaning up.

```bash
eksctl delete cluster --region ${AWS_REGION} --name  ${AWS_PROJECT_NAME}
```
