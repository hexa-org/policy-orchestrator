package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type azureKeyFile struct {
	Tenant string `json:"tenant"`
}

type azureProvider struct {
}

func (p azureProvider) detect(provider string) bool {
	return provider == "azure"
}

func (p azureProvider) name(key []byte) (string, error) {
	var foundKeyFile azureKeyFile
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
	if err != nil || foundKeyFile.Tenant == "" {
		return "", errors.New("unable to read key file, missing tenant")
	}
	return fmt.Sprintf("tenant:%s", foundKeyFile.Tenant), nil
}
