package awscommon

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	logger "golang.org/x/exp/slog"
	"net/http"
)

// AWSHttpClient
// Contents copied from policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon/amazon_http_client.go
type AWSHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

//type AWSClientOptions struct {
//	HTTPClient AWSHttpClient
//	DisableRetry bool
//}

func GetAwsClientConfig(key []byte, httpClient AWSHttpClient) (aws.Config, error) {
	var awsCredentials credentialsInfo
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&awsCredentials)
	if err != nil {
		logger.Error("GetAwsClientConfig msg", "error decode awsCredentials key", "error", err)
		return aws.Config{}, err
	}

	awsOptions := []func(options *config.LoadOptions) error{
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{AccessKeyID: awsCredentials.AccessKeyID, SecretAccessKey: awsCredentials.SecretAccessKey},
		}),
		config.WithRegion(awsCredentials.Region),
	}

	if httpClient != nil {
		awsOptions = append(awsOptions, config.WithHTTPClient(httpClient))
		awsOptions = append(awsOptions, config.WithRetryer(func() aws.Retryer { return aws.NopRetryer{} }))
	}

	//if opt.DisableRetry {
	//	awsOptions = append(awsOptions, config.WithRetryer(func() aws.Retryer { return aws.NopRetryer{} }))
	//}

	return config.LoadDefaultConfig(context.Background(), awsOptions...)
}

type credentialsInfo struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}
