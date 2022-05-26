mkdir -p .freshcloud
cat <<EOF > .freshcloud/concourse-values.yaml
concourse:
  worker:
    replicaCount: 4
  web:
    externalUrl: https://ci.{{index . "DOMAIN"}}
    auth:
      mainTeam:
        localUser: "admin"
secrets:
  localUsers: "admin:{{index . "PASSWORD"}}"
web:
  env:
  ingress:
    enabled: true
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
      kubernetes.io/ingress.class: contour
      ingress.kubernetes.io/force-ssl-redirect: "true"
      projectcontour.io/websocket-routes: "/"
      kubernetes.io/tls-acme: "true"
    hosts:
      - ci.{{index . "DOMAIN"}}
    tls:
      - hosts:
          - ci.{{index . "DOMAIN"}}
        secretName: concourse-cert
EOF
kubectl create namespace concourse
helm repo add concourse https://concourse-charts.storage.googleapis.com/
helm install concourse concourse/concourse -f .freshcloud/concourse-values.yaml -n concourse
if [ $? != 0 ]; then
  echo "Failed to install Concourse. Bummer"
  exit 1
fi
kubectl wait --for=condition=Ready pods --timeout=900s --all -n concourse
echo "Remove concourse by running - kubectl delete ns concourse"