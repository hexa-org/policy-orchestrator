package azurecommon

type AzureKey struct {
	AppId        string `json:"appId"`
	Secret       string `json:"secret"`
	Tenant       string `json:"tenant"`
	Subscription string `json:"subscription"`
}
