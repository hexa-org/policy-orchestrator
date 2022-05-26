fly login -c https://ci.{{index . "REGISTRY_DOMAIN"}} -u admin -p {{index . "REGISTRY_PASSWORD"}} -t {{index . "REGISTRY_CLUSTER_NAME"}}
echo y | fly -t {{index . "REGISTRY_CLUSTER_NAME"}} dp -p build-{{index . "APP_NAME"}}
kubectl delete ns {{index . "APP_NAME"}}