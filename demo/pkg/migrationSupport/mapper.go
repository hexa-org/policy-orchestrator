package migrationSupport

import (
	"strings"

	"github.com/hexa-org/policy-mapper/sdk"
)

func MapSdkProviderName(legacyName string) string {
	switch strings.ToLower(legacyName) {
	case "azure", "azure_apim":
		return sdk.ProviderTypeAzure
	case "amazon":
		return sdk.ProviderTypeCognito
	case "google_cloud", "gcp":
		return sdk.ProviderTypeGoogleCloudIAP
	case "open_policy_agent":
		return sdk.ProviderTypeOpa
	case "noop":
		return "noop"
	}
	return legacyName
}
