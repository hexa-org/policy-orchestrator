package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hexa-org/policy-mapper/sdk"
)

type awsApiGatewayProvider struct {
}

func (p awsApiGatewayProvider) detect(provider string) bool {
	return provider == sdk.ProviderTypeAwsApiGW
}

func (p awsApiGatewayProvider) name(key []byte) (string, error) {
	var foundKeyFile amazonKeyFile
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
	if err != nil || foundKeyFile.Region == "" {
		return "", errors.New("unable to read key file, missing region")
	}
	return fmt.Sprintf("region:%s", foundKeyFile.Region), nil
}
