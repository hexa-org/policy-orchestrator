package amazonwebservices

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"net/http"
)

type AWSHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AWSClientOptions struct {
	HTTPClient   AWSHttpClient
	DisableRetry bool
}

func GetAwsClientConfig(key []byte, opt AWSClientOptions) (aws.Config, error) {
	var awsCredentials credentialsInfo
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&awsCredentials)
	if err != nil {
		return aws.Config{}, err
	}

	awsOptions := []func(options *config.LoadOptions) error{
		config.WithCredentialsProvider(awscredentials.StaticCredentialsProvider{
			Value: aws.Credentials{AccessKeyID: awsCredentials.AccessKeyID, SecretAccessKey: awsCredentials.SecretAccessKey},
		}),
		config.WithRegion(awsCredentials.Region),
	}

	if opt.HTTPClient != nil {
		awsOptions = append(awsOptions, config.WithHTTPClient(opt.HTTPClient))
	}

	if opt.DisableRetry {
		awsOptions = append(awsOptions, config.WithRetryer(func() aws.Retryer { return aws.NopRetryer{} }))
	}

	return config.LoadDefaultConfig(context.Background(), awsOptions...)
}

type credentialsInfo struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}
