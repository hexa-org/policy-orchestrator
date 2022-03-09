package openpolicyagent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"log"
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
	dir := os.TempDir()
	rego, err := o.Client.GetExpressionFromBundle(foundCredentials.BundleUrl, dir)
	if err != nil {
		log.Printf("open-policy-agent, unable to read expression file. %s\n", err)
		return nil, err
	}
	return o.Service.ReadPolicies(bytes.NewReader(rego))
}

func (o *OpaProvider) SetPolicyInfo(integration provider.IntegrationInfo, app provider.ApplicationInfo, policy provider.PolicyInfo) error {
	return nil
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
