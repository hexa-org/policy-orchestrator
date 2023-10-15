FROM openpolicyagent/opa:latest
ADD config/aws-s3-opa-config.yaml /aws-s3-opa-config.yaml
ADD config/github-opa-config.yaml /github-opa-config.yaml
ADD config/gcp-opa-config.yaml /gcp-opa-config.yaml
ADD config/config.yaml /config.yaml
ENTRYPOINT ["/opa"]
CMD ["run"]
