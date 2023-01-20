package openpolicyagent

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type BundleClient interface {
	GetDataFromBundle(path string) ([]byte, error)
	PostBundle(bundle []byte) (int, error)
}

type OpaProvider struct {
	BundleClientOverride BundleClient
	ResourcesDirectory   string
}

func (o *OpaProvider) Name() string {
	return "open_policy_agent"
}

func (o *OpaProvider) DiscoverApplications(info orchestrator.IntegrationInfo) ([]orchestrator.ApplicationInfo, error) {
	c, err := o.credentials(info.Key)
	if err != nil {
		return nil, err
	}

	var apps []orchestrator.ApplicationInfo
	if strings.EqualFold(info.Name, o.Name()) {
		apps = append(apps, orchestrator.ApplicationInfo{
			ObjectID:    c.objectID(),
			Name:        c.ProjectID,
			Description: "Open Policy Agent bundle",
			Service:     "Hexa OPA",
		})
	}
	return apps, nil
}

type Policies struct {
	Policies []Policy `json:"policies"`
}

type Policy struct {
	Meta    Meta     `json:"meta"`
	Actions []Action `json:"actions"`
	Subject Subject  `json:"subject"`
	Object  Object   `json:"object"`
}

type Meta struct {
	Version string `json:"version"`
}

type Action struct {
	ActionUri string `json:"action_uri"`
}

type Subject struct {
	Members []string `json:"members"`
}

type Object struct {
	ResourceID string `json:"resource_id"`
}

func (o *OpaProvider) GetPolicyInfo(integration orchestrator.IntegrationInfo, appInfo orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	key := integration.Key
	client, err := o.ConfigureClient(key)
	if err != nil {
		log.Printf("open-policy-agent, unable to build client: %s", err)
		return nil, fmt.Errorf("invalid client: %w", err)
	}
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(os.TempDir(), fmt.Sprintf("/test-bundle-%d", rand.Uint64()))
	data, err := client.GetDataFromBundle(path)
	if err != nil {
		log.Printf("open-policy-agent, unable to read expression file. %s\n", err)
		return nil, err
	}

	var policies Policies
	unmarshalErr := json.Unmarshal(data, &policies)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	var hexaPolicies []policysupport.PolicyInfo
	for _, p := range policies.Policies {
		var actions []policysupport.ActionInfo
		for _, a := range p.Actions {
			actions = append(actions, policysupport.ActionInfo{ActionUri: a.ActionUri})
		}
		hexaPolicies = append(hexaPolicies, policysupport.PolicyInfo{
			Meta:    policysupport.MetaInfo{Version: p.Meta.Version},
			Actions: actions,
			Subject: policysupport.SubjectInfo{
				Members: p.Subject.Members,
			},
			Object: policysupport.ObjectInfo{
				ResourceID: appInfo.ObjectID, // todo - for now, ensures the correct resource identifier
			},
		})
	}
	return hexaPolicies, nil
}

func (o *OpaProvider) SetPolicyInfo(integration orchestrator.IntegrationInfo, appInfo orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) (int, error) {
	validate := validator.New() // todo - move this up?
	errApp := validate.Struct(appInfo)
	if errApp != nil {
		return http.StatusInternalServerError, fmt.Errorf("invalid app info: %w", errApp)
	}
	errPolicies := validate.Var(policyInfos, "omitempty,dive")
	if errPolicies != nil {
		return http.StatusInternalServerError, fmt.Errorf("invalid policy info: %w", errPolicies)
	}

	key := integration.Key
	client, err := o.ConfigureClient(key)
	if err != nil {
		log.Printf("open-policy-agent, unable to build client: %s", err)
		return http.StatusInternalServerError, fmt.Errorf("invalid client: %w", err)
	}

	var policies []Policy
	for _, p := range policyInfos {
		var actions []Action
		for _, a := range p.Actions {
			actions = append(actions, Action{a.ActionUri})
		}
		policies = append(policies, Policy{
			Meta:    Meta{Version: p.Meta.Version},
			Actions: actions,
			Subject: Subject{
				p.Subject.Members,
			},
			Object: Object{
				ResourceID: appInfo.ObjectID, // todo - for now, ensures the correct resource identifier
			},
		})
	}
	data, marshalErr := json.Marshal(Policies{policies})
	if marshalErr != nil {
		log.Printf("open-policy-agent, unable to create data file. %s\n", marshalErr)
		return http.StatusInternalServerError, marshalErr
	}

	bundle, copyErr := o.MakeDefaultBundle(data)
	if copyErr != nil {
		log.Printf("open-policy-agent, unable to create default bundle. %s\n", copyErr)
		return http.StatusInternalServerError, copyErr
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf("unable to set policy: %v", err)
		}
	}()
	return client.PostBundle(bundle.Bytes())
}

func (o *OpaProvider) MakeDefaultBundle(data []byte) (bytes.Buffer, error) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles/bundle")
	manifest, _ := ioutil.ReadFile(filepath.Join(join, "/.manifest"))
	rego, _ := ioutil.ReadFile(filepath.Join(join, "/policy.rego"))

	// todo - ignoring errors for the moment while spiking

	tempDir := os.TempDir()
	_ = os.Mkdir(filepath.Join(tempDir, "/bundles"), 0744)
	_ = os.Mkdir(filepath.Join(tempDir, "/bundles/bundle"), 0744)
	_ = ioutil.WriteFile(filepath.Join(tempDir, "/bundles/bundle/.manifest"), manifest, 0644)
	_ = ioutil.WriteFile(filepath.Join(tempDir, "/bundles/bundle/data.json"), data, 0644)
	_ = ioutil.WriteFile(filepath.Join(tempDir, "/bundles/bundle/policy.rego"), rego, 0644)

	tar, _ := compressionsupport.TarFromPath(filepath.Join(tempDir, "/bundles"))
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	return buffer, nil
}

type credentials struct {
	ProjectID string          `json:"project_id,omitempty"`
	BundleUrl string          `json:"bundle_url"`
	CACert    string          `json:"ca_cert,omitempty"`
	GCP       *gcpCredentials `json:"gcp,omitempty"`
	AWS       *awsCredentials `json:"aws,omitempty"`
}

func (c credentials) objectID() string {
	if c.GCP != nil {
		return c.GCP.BucketName
	}
	return base64.StdEncoding.EncodeToString([]byte(c.BundleUrl))
}

type gcpCredentials struct {
	BucketName string          `json:"bucket_name,omitempty"`
	ObjectName string          `json:"object_name,omitempty"`
	Key        json.RawMessage `json:"key,omitempty"`
}

type awsCredentials gcpCredentials

func (o *OpaProvider) credentials(key []byte) (credentials, error) {
	var foundCredentials credentials
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundCredentials)
	if err != nil {
		return credentials{}, fmt.Errorf("invalid integration key: %w", err)
	}
	if foundCredentials.ProjectID == "" {
		foundCredentials.ProjectID = "package authz"
	}
	return foundCredentials, nil
}

func (o *OpaProvider) ConfigureClient(key []byte) (BundleClient, error) {
	// todo - do we need ResourcesDirectory here? Are we using it?
	if o.ResourcesDirectory == "" {
		_, file, _, _ := runtime.Caller(0)
		o.ResourcesDirectory = filepath.Join(file, "../resources")
	}

	creds, err := o.credentials(key)
	if err != nil {
		return nil, err
	}

	if creds.GCP != nil {
		return NewGCPBundleClient(
			creds.GCP.BucketName,
			creds.GCP.ObjectName,
			creds.GCP.Key,
		)
	}

	if creds.AWS != nil {
		return NewAWSBundleClient(
			creds.AWS.BucketName,
			creds.AWS.ObjectName,
			creds.AWS.Key,
			AWSBundleClientOptions{},
		)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	if creds.CACert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(creds.CACert))
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
	}

	if o.BundleClientOverride != nil {
		return o.BundleClientOverride, nil
	}

	return &HTTPBundleClient{
		BundleServerURL: creds.BundleUrl,
		HttpClient:      client,
	}, nil
}
