package admin

import (
	"bytes"
	"encoding/json"
	"errors"
)

type bundleFile struct {
	BundleUrl string `json:"bundle_url"`
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
	return "bundle:open-policy-agent", nil
}
