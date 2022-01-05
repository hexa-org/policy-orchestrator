gcloud builds submit --pack image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1,builder=heroku/buildpacks:20 --quiet

gcloud run deploy ${GCP_PROJECT_NAME}-policy-admin \
    --command="admin" \
    --region=${GCP_PROJECT_REGION} \
    --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1

gcloud run deploy ${GCP_PROJECT_NAME}-policy-orchestrator \
    --command="orchestrator" \
    --region=${GCP_PROJECT_REGION} \
    --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1 \
    --ingress internal

gcloud run deploy ${GCP_PROJECT_NAME}-demo \
    --command="demo" \
    --region=${GCP_PROJECT_REGION} \
    --image=gcr.io/${GCP_PROJECT_ID}/${GCP_PROJECT_NAME}:tag1
