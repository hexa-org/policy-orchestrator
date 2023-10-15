package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type bundleFile struct {
	BundleUrl string `json:"bundle_url"`
	ProjectID string `json:"project_id,omitempty"`
	GCP       any    `json:"gcp,omitempty"`
	AWS       any    `json:"aws,omitempty"`
	GITHUB    any    `json:"github,omitempty"`
}

func (b bundleFile) isCloudHosted() bool {
	return b.GCP != nil || b.AWS != nil || b.GITHUB != nil
}

type opaProvider struct {
}

func (p opaProvider) detect(provider string) bool {
	return provider == "open_policy_agent"
}

func (p opaProvider) name(key []byte) (string, error) {
	var foundBundleFile bundleFile
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundBundleFile)
	if err != nil || (foundBundleFile.BundleUrl == "" && !foundBundleFile.isCloudHosted()) {
		return "", errors.New("unable to read key file, missing bundle url")
	}
	projectID := "bundle"
	if foundBundleFile.ProjectID != "" {
		projectID = foundBundleFile.ProjectID
	}
	return fmt.Sprintf("%s:open-policy-agent", projectID), nil
}
