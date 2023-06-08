package openpolicyagent

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"

	"net/http"
	"os"
	"path/filepath"
)

type AWSBundleClient struct {
	bucketName string
	objectName string
	httpClient *s3.Client
}

func NewAWSBundleClient(bucketName, objectName string, key []byte, opts amazonwebservices.AWSClientOptions) (*AWSBundleClient, error) {
	if len(bucketName) == 0 || len(objectName) == 0 {
		return nil, fmt.Errorf("required config: bucket_name, object_name")
	}

	cfg, err := amazonwebservices.GetAwsClientConfig(key, opts)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)

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
