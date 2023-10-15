# Deploy to Azure

## Install Azure CLI

Install the Azure CLI following [these instructions](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) or with
[Homebrew](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-macos)

## Azure Setup

Log in to azure CLI.

```bash
az login
```

Create a `.env_azure.sh` file to store your azure environment variables.

```bash
export APP_NAME=<app name>
export AZ_RESOURCE_GROUP=<resource group>
export AZ_LOCATION=<location>
expoer AZ_AD_TENANT_ID=<tenant id>
export AZ_AKS_CLUSTER_NAME=<cluster name>
export AZ_ACR_NAME=<name>
```

Source the `.env_azure.sh` file.

```bash
source .env_azure.sh
```

A resource group may also be needed.

```bash
az group create --name ${AZ_RESOURCE_GROUP} \
  --location ${AZ_LOCATION}
```

## Build and push images

Build the hexa image.

```bash
pack build hexa --builder heroku/buildpacks:20
```

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
docker tag hexa ${AZ_ACR_NAME}.azurecr.io/hexa:tag1
docker push ${AZ_ACR_NAME}.azurecr.io/hexa:tag1
```

Build an OPA server with configuration via Docker.

From the `./deployments/opa-server` directory run the below commands.

```bash
docker pull openpolicyagent/opa:latest
docker build -t hexa-opa-server:latest .
```

Push image to ACR.

```bash
docker tag hexa-opa-server:latest ${AZ_ACR_NAME}.azurecr.io/hexa-opa-server:latest
docker push ${AZ_ACR_NAME}.azurecr.io/hexa-opa-server:latest
```

## Deploy via App Services

Create App Service Plan.

```bash
az appservice plan create --name ${APP_NAME}plan \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --is-linux
```

Deploy the Hexa Demo App.

```bash
az webapp create --name ${APP_NAME} \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --plan ${APP_NAME}plan \
  --startup-file="demo" \
  --deployment-container-image-name ${AZ_ACR_NAME}.azurecr.io/hexa:tag1

az webapp config appsettings set --name ${APP_NAME} \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --settings PORT=8886

az webapp restart --name ${APP_NAME} \
  --resource-group ${AZ_RESOURCE_GROUP}
```

Deploy the Hexa Demo Config App.

```bash
az webapp create --name ${APP_NAME}-config \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --plan ${APP_NAME}plan \
  --startup-file="democonfig" \
  --deployment-container-image-name ${AZ_ACR_NAME}.azurecr.io/hexa:tag1

az webapp config appsettings set --name ${APP_NAME}-config \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --settings PORT=8889

az webapp restart --name ${APP_NAME}-config \
  --resource-group ${AZ_RESOURCE_GROUP}
```

Deploy the OPA Server.

```bash
az webapp create --name ${APP_NAME}-opa-server \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --plan ${APP_NAME}plan \
  --startup-file="run --server --addr 0.0.0.0:8887 --config-file /config.yaml" \
  --deployment-container-image-name ${AZ_ACR_NAME}.azurecr.io/hexa-opa-server:latest

az webapp config appsettings set --name ${APP_NAME}-opa-server \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --settings PORT=8887

az webapp config appsettings set --name ${APP_NAME}-opa-server \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --settings HEXA_DEMO_CONFIG_URL=https://$(az webapp show --name ${APP_NAME}-config --resource-group ${AZ_RESOURCE_GROUP} | jq -r '.defaultHostName')

az webapp restart --name ${APP_NAME}-opa-server \
  --resource-group ${AZ_RESOURCE_GROUP}
```

Update the Hexa demo config.

```bash
az webapp config appsettings set --name ${APP_NAME} \
  --resource-group ${AZ_RESOURCE_GROUP} \
  --settings OPA_SERVER_URL=https://$(az webapp show --name ${APP_NAME}-opa-server --resource-group ${AZ_RESOURCE_GROUP} | jq -r '.defaultHostName')/v1/data/authz/allow

az webapp restart --name ${APP_NAME} \
  --resource-group ${AZ_RESOURCE_GROUP}
```

## Create AD App Registration

_Below is work in progress_

Here is a [link](https://www.shawntabrizi.com/aad/common-microsoft-resources-azure-active-directory)
describing the required-resource-accesses file resourceAccess and resourceAppId are specific to associated apis
look for User.Read (az ad sp list | grep User.Read)

Create an Azure Active Directory app.

```bash
az ad app create \
  --display-name ${APP_NAME} \
  --web-home-page-url "https://${APP_NAME}.azurewebsites.net" \
  --web-redirect-uris "https://${APP_NAME}.azurewebsites.net/.auth/login/aad/callback" \
  --enable-id-token-issuance \
  --sign-in-audience "AzureADandPersonalMicrosoftAccount" \
  --required-resource-accesses @required-resource-accesses.json.txt
```

Enable webapp authentication and authorization for the demo app.

```bash
export AD_APP_ID=$(az ad app list --filter "displayname eq '${APP_NAME}'" | jq -r '.[].appId')
echo "Newly created ad app with id ${APP_NAME}"

az ad app credential reset --id ${AD_APP_ID}
export AZ_APP_SECRET=...

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

echo "Creating the service principal for the ${APP_NAME} app"
az ad sp create --id ${AD_APP_ID}

export AD_SP_ID=$(az ad sp list --all --query "[?appId=='$AD_APP_ID']" | jq -r '.[].id')
echo "Newly created service principal with id ${AD_SP_ID}"

echo "Updating the service principal for the ${APP_NAME} app"
az ad sp update --id ${AD_SP_ID} --set "appRoleAssignmentRequired=true"
az ad sp update --id ${AD_SP_ID} --set "tags=[\"WindowsAzureActiveDirectoryIntegratedApp\"]"

echo "Deleting the azure ad app for ${APP_NAME}"
az ad app delete --id $(az ad app list --filter "displayname eq '${APP_NAME}'" | jq -r '.[].appId')
```
