ARG GCP_PROJECT_ID
FROM gcr.io/${GCP_PROJECT_ID}/opa:latest
ADD config/config.yaml /config.yaml
ENTRYPOINT ["/opa"]
CMD ["run"]
