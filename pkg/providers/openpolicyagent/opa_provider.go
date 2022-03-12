package openpolicyagent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type OpaProvider struct {
	Client  BundleClient
	Service OpaService
}

func (o *OpaProvider) Name() string {
	return "open_policy_agent"
}

func (o *OpaProvider) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo, err error) {
	c := o.credentials(info.Key)
	if strings.EqualFold(info.Name, o.Name()) {
		apps = append(apps, provider.ApplicationInfo{
			ObjectID:    base64.StdEncoding.EncodeToString([]byte(c.BundleUrl)),
			Name:        "package authz",
			Description: "Open policy agent bundle",
		})
	}
	return apps, err
}

func (o *OpaProvider) GetPolicyInfo(integration provider.IntegrationInfo, _ provider.ApplicationInfo) ([]provider.PolicyInfo, error) {
	o.ensureClientIsAvailable()
	key := integration.Key
	foundCredentials := o.credentials(key)
	path := filepath.Join(os.TempDir(), fmt.Sprintf("/test-bundle-%d", rand.Uint64()))
	rego, err := o.Client.GetExpressionFromBundle(foundCredentials.BundleUrl, path)
	if err != nil {
		log.Printf("open-policy-agent, unable to read expression file. %s\n", err)
		return nil, err
	}
	return o.Service.ReadPolicies(bytes.NewReader(rego))
}

func (o *OpaProvider) SetPolicyInfo(integration provider.IntegrationInfo, _ provider.ApplicationInfo, policy provider.PolicyInfo) error {
	o.ensureClientIsAvailable()
	key := integration.Key
	foundCredentials := o.credentials(key)
	var rego bytes.Buffer
	writeErr := o.Service.WritePolicies([]provider.PolicyInfo{policy}, &rego)
	if writeErr != nil {
		log.Printf("open-policy-agent, unable to write expression file. %s\n", writeErr)
		return writeErr
	}

	bundle, copyErr := o.MakeDefaultBundle(rego.Bytes())
	if copyErr != nil {
		log.Printf("open-policy-agent, unable to create default bundle. %s\n", copyErr)
		return copyErr
	}
	return o.Client.PostBundle(foundCredentials.BundleUrl, bundle.Bytes())
}

func (o *OpaProvider) MakeDefaultBundle(rego []byte) (bytes.Buffer, error) {
	_, file, _, _ := runtime.Caller(0)
	join := filepath.Join(file, "../resources/bundles/bundle")
	manifest, _ := ioutil.ReadFile(filepath.Join(join, "/.manifest"))
	data, _ := ioutil.ReadFile(filepath.Join(join, "/data.json"))

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

func (o *OpaProvider) ensureClientIsAvailable() {
	if o.Client.HttpClient == nil {
		o.Client = BundleClient{HttpClient: &http.Client{}}
	}
	if o.Service.ResourcesDirectory == "" {
		_, file, _, _ := runtime.Caller(0)
		resourcesDirectory := filepath.Join(file, "../resources")
		o.Service = OpaService{resourcesDirectory}
	}
}
