package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hexa-org/policy-mapper/sdk"
)

type googleKeyFile struct {
	ProjectId string `json:"project_id"`
}

type googleProvider struct {
}

func (p googleProvider) detect(provider string) bool {
	return provider == sdk.ProviderTypeGoogleCloudIAP || provider == sdk.ProviderTypeGoogleCloudLegacy
}

func (p googleProvider) name(key []byte) (string, error) {
	var foundKeyFile googleKeyFile
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
	if err != nil {
		return "", err
	}
	if foundKeyFile.ProjectId == "" {
		return "", errors.New("unable to read key file, missing project")
	}
	return fmt.Sprintf("project:%s", foundKeyFile.ProjectId), nil
}
