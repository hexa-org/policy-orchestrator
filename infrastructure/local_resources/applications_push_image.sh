curl --user "admin:{{index . "REGISTRY_PASSWORD"}}" -X POST \
  https://registry.{{index . "REGISTRY_DOMAIN"}}/api/v2.0/projects \
  -H "Content-type: application/json" --data \
  '{ "project_name": "'{{index . "APP_NAME"}}'", "metadata":
   { "auto_scan": "true", "enable_content_trust":
     "false", "prevent_vul": "false", "public":
     "true", "reuse_sys_cve_whitelist": "true",
     "severity": "high" }
   }'
docker login -u admin -p {{index . "REGISTRY_PASSWORD"}} https://registry.{{index . "REGISTRY_DOMAIN"}}
docker tag {{index . "APP_IMAGE_NAME"}} registry.{{index . "REGISTRY_DOMAIN"}}/{{index . "APP_NAME"}}/{{index . "APP_IMAGE_NAME"}}:latest
docker push registry.{{index . "REGISTRY_DOMAIN"}}/{{index . "APP_NAME"}}/{{index . "APP_IMAGE_NAME"}}:latest