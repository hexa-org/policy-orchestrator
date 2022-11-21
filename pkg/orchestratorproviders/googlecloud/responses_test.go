package googlecloud_test

var appEngineAppsJSON = []byte(`{
  "name": "apps/hexa-demo",
  "id": "hexa-demo",
  "authDomain": "gmail.com",
  "locationId": "us-central",
  "codeBucket": "staging.hexa-demo.appspot.com",
  "servingStatus": "SERVING",
  "defaultHostname": "hexa-demo.uc.r.appspot.com",
  "defaultBucket": "hexa-demo.appspot.com",
  "serviceAccount": "hexa-demo@appspot.gserviceaccount.com",
  "iap": {
    "enabled": true,
    "oauth2ClientId": "oauth2ClientId",
    "oauth2ClientSecretSha256": "oauth2ClientSecretSha256"
  },
  "gcrDomain": "us.gcr.io",
  "databaseType": "CLOUD_DATASTORE_COMPATIBILITY",
  "featureSettings": {
    "splitHealthChecks": true,
    "useContainerOptimizedOs": true
  }
}`)

var backendAppsJSON = []byte(`{
  "id": "projects/aProject/global/backendServices",
  "items": [
    {
      "id": "0000000000000001",
      "name": "k8s1-aName",
      "description": "{\"kubernetes.io/service-name\":\"default/aName\",\"kubernetes.io/service-port\":\"8887\",\"x-features\":[\"NEG\"]}",
      "kind": "compute#backendService"
    },
    {
      "id": "0000000000000002",
      "name": "k8s1-anotherName",
      "description": "{\"kubernetes.io/service-name\":\"default/anotherName\",\"kubernetes.io/service-port\":\"8887\",\"x-features\":[\"NEG\"]}",
      "kind": "compute#backendService"
    },
    {
      "id": "0000000000000002",
      "name": "cloud-run-app",
      "description": "some description",
      "kind": "compute#backendService"
    }
  ],
  "selfLink": "https://www.googleapis.com/compute/v1/projects/aProject/global/backendServices",
  "kind": "compute#backendServiceList"
}`)

var policyJSON = []byte(`{
  "bindings": [
    {
      "role": "roles/resourcemanager.organizationAdmin",
      "members": [
        "user:phil@example.com",
        "group:admins@example.com",
        "domain:google.com",
        "serviceAccount:my-project-id@appspot.gserviceaccount.com"
      ]
    },
    {
      "role": "roles/resourcemanager.organizationViewer",
      "members": [
        "user:eve@example.com"
      ],
      "condition": {
        "title": "expirable access",
        "description": "Does not grant access after Sep 2020",
        "expression": "request.time < timestamp('2020-10-01T00:00:00.000Z')"
      }
    }
  ],
  "etag": "BwWWja0YfJA=",
  "version": 3
}`)

var projectJSON = []byte(`{
  "type": "service_account",
  "project_id": "google-cloud-project-id",
  "private_key_id": "",
  "private_key": "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----\n",
  "client_email": "google-cloud-project-id@google-cloud-project-id.iam.gserviceaccount.com",
  "client_id": "000000000000000000000",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/google-cloud-project-id%google-cloud-project-id.iam.gserviceaccount.com"
}`)
