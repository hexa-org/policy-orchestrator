# Deploy to Azure

Log in to azure CLI.

```bash
az login
```

Create a `.env_azure.sh` file to store your azure environment variables.

```bash
export APP_NAME=<app name>
export AZ_RESOURCE_GROUP=<resource group>
export AZ_AKS_CLUSTER_NAME=<cluster name>
export AZ_ACR_NAME=<name>
```

## Build and push images

Create container registry.

```bash
  az acr create --name ${AZ_ACR_NAME} \
--resource-group ${AZ_RESOURCE_GROUP} \
--sku standard \
--admin-enabled true
```

Log in to azure registry.

```bash
az acr login --name ${AZ_ACR_NAME}
```

Tag and push demo app image.

```bash
docker tag ${APP_NAME} ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1
docker push ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1
```

Build and push OPA Server.

From the `./deployments/opa-server` directory run the below commands.

```bash
docker pull openpolicyagent/opa:latest
docker build -t ${APP_NAME}-opa-server:latest .
```

Push image to ACR.

```bash
docker tag ${APP_NAME}-opa-server:latest ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}-opa-server:latest
docker push ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}-opa-server:latest
```

## Deploy to App Services

Create App Service Plan.

```bash
az appservice plan create --name ${APP_NAME}plan \
--resource-group ${AZ_RESOURCE_GROUP} \
--is-linux
```

Deploy Hexa Demo App.

```bash
az webapp create --name ${APP_NAME} \
--resource-group ${AZ_RESOURCE_GROUP} \
--plan ${APP_NAME}plan \
--startup-file="demo" \
--deployment-container-image-name ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1

az webapp config appsettings set --name ${APP_NAME} \
--resource-group ${AZ_RESOURCE_GROUP} \
--settings PORT=8881

az webapp config container set --name ${APP_NAME} \
--resource-group ${AZ_RESOURCE_GROUP} \
--docker-custom-image-name ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1 \
--docker-registry-server-url "https://${AZ_ACR_NAME}.azurecr.io"

az webapp restart --name ${APP_NAME} \
--resource-group ${AZ_RESOURCE_GROUP}

az webapp show --name ${APP_NAME} \
--resource-group ${AZ_RESOURCE_GROUP} \
| jq -r '.defaultHostName'
```

Deploy OPA Server.

```bash
az webapp create --name ${APP_NAME}-demo-opa-agent \
--resource-group ${AZ_RESOURCE_GROUP} \
--plan ${APP_NAME}plan \
--startup-file="run --server --addr 0.0.0.0:8881 --config-file /config.yaml" \
--deployment-container-image-name ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}-opa-server:latest

az webapp config appsettings set --name ${APP_NAME}-demo-opa-agent \
--resource-group ${AZ_RESOURCE_GROUP} \
--settings PORT=8881

az webapp config container set --name ${APP_NAME}-demo-opa-agent \
--resource-group ${AZ_RESOURCE_GROUP} \
--docker-custom-image-name ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}-opa-server:latest \
--docker-registry-server-url "https://${AZ_ACR_NAME}.azurecr.io"

az webapp restart --name ${APP_NAME}-demo-opa-agent \
--resource-group ${AZ_RESOURCE_GROUP}

az webapp show --name ${APP_NAME}-demo-opa-agent \
--resource-group ${AZ_RESOURCE_GROUP} \
| jq -r '.defaultHostName'
```

Update config for both apps.

```bash
opa_url=$(az webapp show --name ${APP_NAME}-demo-opa-agent \
--resource-group ${AZ_RESOURCE_GROUP} \
| jq -r '.defaultHostName')

hexa_demo_url=$(az webapp show --name ${APP_NAME}-demo \
--resource-group ${AZ_RESOURCE_GROUP} \
| jq -r '.defaultHostName')

az webapp config appsettings set --name ${APP_NAME}-demo \
--resource-group ${AZ_RESOURCE_GROUP} \
--settings OPA_SERVER_URL=https://${opa_url}/v1/data/authz/allow

az webapp config appsettings set --name ${APP_NAME}-demo-opa-agent \
--resource-group ${AZ_RESOURCE_GROUP} \
--settings HEXA_DEMO_URL=https://${hexa_demo_url}
```

Restart both apps.

## Deploy to Kubernetes - AKS

Create cluster.

```bash
az aks create \
    --resource-group ${AZ_RESOURCE_GROUP} \
    --name ${AZ_AKS_CLUSTER_NAME} \
    --node-count 2 \
    --generate-ssh-keys \
    --attach-acr ${AZ_ACR_NAME}
```

View cluster.

```bash
az aks list --resource-group ${AZ_RESOURCE_GROUP}
```

Connect to cluster.

```bash
az aks get-credentials --resource-group ${AZ_RESOURCE_GROUP} --name ${AZ_AKS_CLUSTER_NAME}
```

Deploy demo app objects.

```bash
envsubst < kubernetes/demo/deployment.yaml | kubectl apply -f -
envsubst < kubernetes/demo/service.yaml | kubectl apply -f -
```

Deploy OPA Agent objects.

```bash
envsubst < kubernetes/opa-server/deployment.yaml | kubectl apply -f -
envsubst < kubernetes/opa-server/service.yaml | kubectl apply -f -
```

## Update Webapp Auth

```bash

# Create AD App Registration
az ad app create \
--display-name ${APP_NAME} \
--homepage "https://${APP_NAME}.azurewebsites.net" \
--reply-urls "https://${APP_NAME}.azurewebsites.net/.auth/login/aad/callback" \
--available-to-other-tenants false \
--required-resource-accesses @required-resource-accesses.json.txt \
--password ${AZ_APP_SECRET}

AD_APP_ID=$(az ad app list --filter "displayname eq '${APP_NAME}'" | jq -r '.[].appId')
echo "Newly created ad app with id ${APP_NAME}"

echo "Adding the authV2 extension"
az extension add --name authV2

echo "Updating the ${APP_NAME} app with microsoft auth provider"
az webapp auth microsoft update --name ${APP_NAME} \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --client-id ${AD_APP_ID} \
  --client-secret ${AZ_APP_SECRET} \
  --yes \
  --allowed-audiences  "api://${AD_APP_ID}" \
  --issuer "https://sts.windows.net/${AZ_AD_TENANT_ID}/"

echo "Creating the service principal for the ${AZ_PROJECT_NAME} app"
az ad sp create --id ${AD_APP_ID}

AD_SP_ID=$(az ad sp list --query "[?appId=='$AD_APP_ID']" | jq -r '.[].objectId')
echo "Newly created service principal with id ${AD_SP_ID}"

echo "Updating the service principal for the ${AZ_PROJECT_NAME} app"
az ad sp update --id ${AD_SP_ID} --set "appRoleAssignmentRequired=true" --add tags WindowsAzureActiveDirectoryIntegratedApp
```

Add a user to the app.

```bash
userId=<user id>
appId=$(az ad app list --filter "displayname eq '${APP_NAME}'" | jq -r '.[].appId')
spObjectId=$(az ad sp list --query "[?appId=='$appId']" | jq -r '.[].objectId')
appRoleId='00000000-0000-0000-0000-000000000000'
az rest --method post --uri "https://graph.microsoft.com/beta/users/$userId/appRoleAssignments" \
--body "{'appRoleId': '$appRoleId','principalId': '$userId','resourceId': '$spObjectId'}" \
--headers "Content-Type=application/json"
```

## az CLI Commands

List Web Services apps

```bash
az webapp list --resource-group ${AZ_RESOURCE_GROUP} | jq -r '.[].name'
```

List AD App Registrations

List all

```bash
az ad app list
```

```bash
AD_APP_ID=$(az ad app list --filter "displayname eq '${APP_NAME}'" | jq -r '.[].appId')
az ad app show --id ${AD_APP_ID}
```

List AD Service Principals (Enterprise Apps)

```bash
AD_APP_ID=$(az ad app list --filter "displayname eq '${APP_NAME}'" | jq -r '.[].appId')
AD_SP_ID=$(az ad sp list --query "[?appId=='$AD_APP_ID']" | jq -r '.[].objectId')
az ad sp show --id ${AD_SP_ID}
```

# Making REST Calls to Azure APIs

You need to create an AD App Registration to make API requests. 

1. Create an AD App Registration
2. Create secret for App Registration
3. Enable Microsoft Graph API permissions and choose "Application Permissions"

To make requests to the API you need to exchange your client secret and client ID for an Access Token
and use that token to make Graph API requests.

