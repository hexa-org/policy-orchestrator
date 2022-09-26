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
}

type opaProvider struct {
}

func (p opaProvider) detect(provider string) bool {
	return provider == "open_policy_agent"
}

func (p opaProvider) name(key []byte) (string, error) {
	var foundKeyFile bundleFile
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
	if err != nil || foundKeyFile.BundleUrl == "" {
		return "", errors.New("unable to read key file, missing bundle url")
	}
	projectID := "bundle"
	if foundKeyFile.ProjectID != "" {
		projectID = foundKeyFile.ProjectID
	}
	return fmt.Sprintf("%s:open-policy-agent", projectID), nil
}
