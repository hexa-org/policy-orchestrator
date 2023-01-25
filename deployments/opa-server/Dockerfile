FROM openpolicyagent/opa:latest
ADD config/github-opa-config.yaml /github-opa-config.yaml
ADD config/gcp-opa-config.yaml /gcp-opa-config.yaml
ADD config/config.yaml /config.yaml
ENTRYPOINT ["/opa"]
CMD ["run"]
