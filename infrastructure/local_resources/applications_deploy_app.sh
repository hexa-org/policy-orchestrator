mkdir -p .freshcloud
sha=$(curl -s --user "admin:{{index . "REGISTRY_PASSWORD"}}" -X GET \
  "https://registry.{{index . "REGISTRY_DOMAIN"}}/api/v2.0/projects/{{index . "APP_NAME"}}/repositories/{{index . "APP_IMAGE_NAME"}}/artifacts" \
  | jq -r '.[].digest'|head -1)
export IMAGE="registry.{{index . "REGISTRY_DOMAIN"}}/{{index . "APP_NAME"}}/{{index . "APP_IMAGE_NAME"}}@${sha}"
echo "Found image ${IMAGE}"
kubectl create namespace {{index . "APP_NAME"}}
envsubst < {{index . "APP_CONFIGURATION_PATH"}} > .freshcloud/{{index . "APP_NAME"}}.yaml
kubectl apply -f .freshcloud/{{index . "APP_NAME"}}.yaml
echo "Deploy {{index . "APP_NAME"}} to https://{{index . "APP_NAME"}}.{{index . "DOMAIN"}}"
echo "Remove the app by running - kubectl delete ns {{index . "APP_NAME"}}"