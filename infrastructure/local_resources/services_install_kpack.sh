kubectl create namespace kpack
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
kubectl apply -f https://github.com/pivotal/kpack/releases/download/v0.5.1/release-0.5.1.yaml
if [ $? != 0 ]; then
  echo "Failed to install Kpack. Bummer"
  exit 1
fi
kubectl wait --for=condition=Ready pods --timeout=900s --all -n kpack
REGISTRY="registry.{{index .DOMAIN}}"
cat <<EOF | kubectl apply -f -
apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  name: base
spec:
  id: "heroku-20"
  buildImage:
    image: "heroku/pack:20-build"
  runImage:
    image: "heroku/pack:20"
EOF
cat <<EOF | kubectl apply -f -
apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  name: default
spec:
  sources:
  - image: heroku/buildpacks:20
EOF
kubectl create secret docker-registry ${REGISTRY} \
  --docker-username=admin \
  --docker-password={{index . "PASSWORD"}} \
  --docker-server=https://${REGISTRY}/ \
  --namespace default
echo "Remove kpack by running - kubectl delete ns kpack"