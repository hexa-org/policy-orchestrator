package migrationSupport

import (
	"testing"

	"github.com/hexa-org/policy-mapper/sdk"
)

func TestMapSdkProviderName(t *testing.T) {
	type args struct {
		legacyName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Unknown ignored", args: args{legacyName: "unknown"}, want: "unknown"},
		{name: "azure test", args: args{legacyName: "azure"}, want: sdk.ProviderTypeAzure},
		{name: "amazon test", args: args{legacyName: "amazon"}, want: sdk.ProviderTypeCognito},
		{name: "google test", args: args{legacyName: "google_cloud"}, want: sdk.ProviderTypeGoogleCloudIAP},
		{name: "opa test", args: args{legacyName: "open_policy_agent"}, want: sdk.ProviderTypeOpa},
		{name: "noop test", args: args{legacyName: "noop"}, want: "noop"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapSdkProviderName(tt.args.legacyName); got != tt.want {
				t.Errorf("MapSdkProviderName() = %v, want %v", got, tt.want)
			}
		})
	}
}
