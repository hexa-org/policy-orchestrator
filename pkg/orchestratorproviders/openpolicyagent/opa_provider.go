package openpolicyagent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type OpaProvider struct {
	BundleClientOverride BundleClient
	ResourcesDirectory   string
}

func (o *OpaProvider) Name() string {
	return "open_policy_agent"
}

func (o *OpaProvider) DiscoverApplications(info orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	c := o.credentials(info.Key)
	if strings.EqualFold(info.Name, o.Name()) {
		apps = append(apps, orchestrator.ApplicationInfo{
			ObjectID:    base64.StdEncoding.EncodeToString([]byte(c.BundleUrl)),
			Name:        "package authz",
			Description: "Open policy agent bundle",
		})
	}
	return apps, err
}

type Wrapper struct {
	Policies []Policy `json:"policies"`
}

type Policy struct {
	Version string  `json:"version"`
	Action  string  `json:"action"`
	Subject Subject `json:"subject"`
	Object  Object  `json:"object"`
}

type Subject struct {
	AuthenticatedUsers []string `json:"authenticated_users"`
}

type Object struct {
	Resources []string `json:"resources"`
}

func (o *OpaProvider) GetPolicyInfo(integration orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo) ([]policysupport.PolicyInfo, error) {
	client := o.ensureClientIsAvailable()
	key := integration.Key
	foundCredentials := o.credentials(key)
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(os.TempDir(), fmt.Sprintf("/test-bundle-%d", rand.Uint64()))
	data, err := client.GetDataFromBundle(foundCredentials.BundleUrl, path)
	if err != nil {
		log.Printf("open-policy-agent, unable to read expression file. %s\n", err)
		return nil, err
	}

	var policies Wrapper
	unmarshalErr := json.Unmarshal(data, &policies)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	var hexaPolicies []policysupport.PolicyInfo
	for _, p := range policies.Policies {
		hexaPolicies = append(hexaPolicies, policysupport.PolicyInfo{
			Version: p.Version,
			Action:  p.Action,
			Subject: policysupport.SubjectInfo{
				AuthenticatedUsers: p.Subject.AuthenticatedUsers,
			},
			Object: policysupport.ObjectInfo{
				Resources: p.Object.Resources,
			},
		})
	}
	return hexaPolicies, nil
}

func (o *OpaProvider) SetPolicyInfo(integration orchestrator.IntegrationInfo, _ orchestrator.ApplicationInfo, policyInfos []policysupport.PolicyInfo) error {
	client := o.ensureClientIsAvailable()
	key := integration.Key
	foundCredentials := o.credentials(key)

	var policies []Policy
	for _, p := range policyInfos {
		policies = append(policies, Policy{
			Version: p.Version,
			Action:  p.Action,
			Subject: Subject{
				p.Subject.AuthenticatedUsers,
			},
			Object: Object{
				p.Object.Resources,
			},
		})
	}
	data, marshalErr := json.Marshal(Wrapper{policies})
	if marshalErr != nil {
		log.Printf("open-policy-agent, unable to create data file. %s\n", marshalErr)
		return marshalErr
	}

	bundle, copyErr := o.MakeDefaultBundle(data)
	if copyErr != nil {
		log.Printf("open-policy-agent, unable to create default bundle. %s\n", copyErr)
		return copyErr
	}
	return client.PostBundle(foundCredentials.BundleUrl, bundle.Bytes())
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

///

type credentials struct {
	BundleUrl string `json:"bundle_url"`
}

func (o *OpaProvider) credentials(key []byte) credentials {
	var foundCredentials credentials
	_ = json.NewDecoder(bytes.NewReader(key)).Decode(&foundCredentials)
	return foundCredentials
}

func (o *OpaProvider) ensureClientIsAvailable() BundleClient {
	if o.ResourcesDirectory == "" {
		_, file, _, _ := runtime.Caller(0)
		o.ResourcesDirectory = filepath.Join(file, "../resources")
	}

	if o.BundleClientOverride.HttpClient != nil {
		return o.BundleClientOverride
	}
	return BundleClient{HttpClient: &http.Client{}}
}
