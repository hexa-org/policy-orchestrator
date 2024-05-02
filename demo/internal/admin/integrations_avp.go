package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hexa-org/policy-mapper/sdk"
)

type avpKeyFile struct {
	Region string `json:"region"`
}

type avpProvider struct {
}

func (p avpProvider) detect(provider string) bool {
	return provider == sdk.ProviderTypeAvp
}

func (p avpProvider) name(key []byte) (string, error) {
	var foundKeyFile avpKeyFile
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
	if err != nil || foundKeyFile.Region == "" {
		return "", errors.New("unable to read key file, missing region")
	}
	return fmt.Sprintf("region:%s", foundKeyFile.Region), nil
}
