package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hexa-org/policy-mapper/sdk"
)

type amazonKeyFile struct {
	Region string `json:"region"`
}

type amazonProvider struct {
}

func (p amazonProvider) detect(provider string) bool {
	return provider == "amazon" || provider == sdk.ProviderTypeCognito
}

func (p amazonProvider) name(key []byte) (string, error) {
	var foundKeyFile amazonKeyFile
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
	if err != nil || foundKeyFile.Region == "" {
		return "", errors.New("unable to read key file, missing region")
	}
	return fmt.Sprintf("region:%s", foundKeyFile.Region), nil
}
