package openpolicyagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"net/http"
	"os"
	"path/filepath"
)

type AWSHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AWSBundleClientOptions struct {
	HTTPClient   AWSHttpClient
	DisableRetry bool
}

type AWSBundleClient struct {
	bucketName string
	objectName string
	httpClient *s3.Client
}

func (c *AWSBundleClientOptions) WithAWSHTTPClient(client AWSHttpClient) {
	c.HTTPClient = client
}

func NewAWSBundleClient(bucketName, objectName string, key []byte, opts AWSBundleClientOptions) (*AWSBundleClient, error) {
	if len(bucketName) == 0 || len(objectName) == 0 {
		return nil, fmt.Errorf("required config: bucket_name, object_name")
	}

	s3Client, err := getS3Client(key, opts)
	if err != nil {
		return nil, err
	}

	bundleClient := &AWSBundleClient{
		bucketName: bucketName,
		objectName: objectName,
		httpClient: s3Client,
	}

	return bundleClient, nil
}

func (a *AWSBundleClient) GetDataFromBundle(path string) ([]byte, error) {
	resp, err := a.httpClient.GetObject(context.Background(),
		&s3.GetObjectInput{Bucket: aws.String(a.bucketName), Key: aws.String(a.objectName)})

	if err != nil {
		return nil, fmt.Errorf("unable to read bundle object from AWS S3 bucket: %w", err)
	}

	defer resp.Body.Close()

	gz, gzipErr := compressionsupport.UnGzip(resp.Body)
	if gzipErr != nil {
		return nil, gzipErr
	}

	tarErr := compressionsupport.UnTarToPath(bytes.NewReader(gz), path)
	if tarErr != nil {
		return nil, tarErr
	}

	return os.ReadFile(filepath.Join(path, "/bundle/data.json"))
}

func (a *AWSBundleClient) PostBundle(bundle []byte) (int, error) {
	_, err := a.httpClient.PutObject(context.Background(),
		&s3.PutObjectInput{
			Bucket:      aws.String(a.bucketName),
			Key:         aws.String(a.objectName),
			Body:        bytes.NewReader(bundle),
			ContentType: aws.String(http.DetectContentType(bundle)),
		})

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to write bundle object to AWS S3 bucket: %w", err)
	}

	return http.StatusCreated, nil
}

func getS3Client(key []byte, opts AWSBundleClientOptions) (*s3.Client, error) {
	var awsCredentials CredentialsInfo
	err := json.NewDecoder(bytes.NewReader(key)).Decode(&awsCredentials)
	if err != nil {
		return nil, err
	}

	awsOptions := []func(options *config.LoadOptions) error{
		config.WithCredentialsProvider(awscredentials.StaticCredentialsProvider{
			Value: aws.Credentials{AccessKeyID: awsCredentials.AccessKeyID, SecretAccessKey: awsCredentials.SecretAccessKey},
		}),
		config.WithRegion(awsCredentials.Region),
	}

	if opts.HTTPClient != nil {
		awsOptions = append(awsOptions, config.WithHTTPClient(opts.HTTPClient))
	}

	if opts.DisableRetry {
		awsOptions = append(awsOptions, config.WithRetryer(func() aws.Retryer { return aws.NopRetryer{} }))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), awsOptions...)

	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

type CredentialsInfo struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}
