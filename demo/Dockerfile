FROM cgr.dev/chainguard/static

LABEL org.opencontainers.image.authors="phil.hunt@independentid.com"
LABEL org.opencontainers.image.source="https://github.com/hexa-org/policy-orhestrator"

WORKDIR /app

# Any non-zero number will do, and unfortunately a named user will not, as k8s
# pod securityContext runAsNonRoot can't resolve the user ID:
# https://github.com/kubernetes/kubernetes/issues/40958.
ARG USER=1000:1000
USER ${USER}

# ADD docker/config/aws-s3-opa-config.yaml ./aws-s3-opa-config.yaml
# ADD docker/config/github-opa-config.yaml ./github-opa-config.yaml
# ADD docker/config/gcp-opa-config.yaml ./gcp-opa-config.yaml
# ADD docker/config/config.yaml ./config.yaml

ADD --chmod=ug+x ./hexaAdminUi ./hexaAdminUi
ADD --chmod=ug+x ./hexaDemo ./hexaDemo
ADD --chmod=ug+x ./hexaOrchestrator ./hexaOrchestrator

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose

# ENTRYPOINT ["hexaOpa run --server"]
# ENTRYPOINT ["/app/hexaOpa"]
EXPOSE 8181:8181

CMD ["/app/hexaOpa","run","--server","--log-level","debug","--addr",":8181"]
# CMD "./hexaOpa help"

