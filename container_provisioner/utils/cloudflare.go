package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	r2 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Obj is an object struct for R2 bucket objects
type R2Obj struct {
	ChecksumAlgorithm string `json:"checksumAlgorithm"`
	Etag              string `json:"Etag"`
	Key               string `json:"Key"`
	LastModified      string `json:"LastModified"`
	Size              int64  `json:"Size"`
	StorageClass      string `json:"StorageClass"`
	Metadata          string
}

// R2Service wraps the R2 client and bucket configuration
type R2Service struct {
	client *r2.Client
	bucket string
	ctx    context.Context
}

// NewR2Service creates a new R2Service from a credentials JSON file
func NewR2Service(credsPath string) (*R2Service, error) {
	data, err := ParseCredsFromJSON(credsPath)
	if err != nil {
		return nil, fmt.Errorf("fail to parse creds from %s: %w", credsPath, err)
	}

	client, err := createR2Client(data.AccessKeyID, data.AccessKeySecret, data.AccountID)
	if err != nil {
		return nil, fmt.Errorf("fail to create R2 client: %w", err)
	}

	return &R2Service{
		client: client,
		bucket: data.BucketName,
		ctx:    context.Background(),
	}, nil
}

// UploadObject uploads an object to R2 and removes the local file
func (s *R2Service) UploadObject(fileName string, uploadIdentifier string, fileData io.Reader) error {
	_, err := s.client.PutObject(s.ctx, &r2.PutObjectInput{
		Bucket: &s.bucket,
		Key:    aws.String(fileName),
		Body:   fileData,
		Metadata: map[string]string{
			"uploadedby": uploadIdentifier,
		},
	})
	if err != nil {
		return fmt.Errorf("fail to upload file %s to R2: %w", fileName, err)
	}

	log.Printf("File: %s uploaded", fileName)

	err = os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("fail to remove the uploaded file %s from local filesystem: %w", fileName, err)
	}

	return nil
}

// ListObjects lists objects in the R2 bucket
func (s *R2Service) ListObjects() ([]R2Obj, error) {
	listObjectsOutput, err := s.client.ListObjectsV2(s.ctx, &r2.ListObjectsV2Input{
		Bucket:     &s.bucket,
		FetchOwner: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("fail to list objects in R2: %w", err)
	}

	files := []R2Obj{}

	for _, object := range listObjectsOutput.Contents {
		obj, err := json.MarshalIndent(object, "", "\t")
		if err != nil {
			return nil, fmt.Errorf("fail to marshal R2 object: %w", err)
		}

		r2Obj := R2Obj{}
		err = json.Unmarshal(obj, &r2Obj)
		if err != nil {
			return nil, fmt.Errorf("fail to unmarshal R2 object: %w", err)
		}

		files = append(files, r2Obj)
	}

	return files, nil
}

// EnrichMetaData enriches the R2 object list with metadata from HeadObject calls
func (s *R2Service) EnrichMetaData(r2ObjectList []R2Obj) ([]R2Obj, error) {
	metaDataMap := []map[string]string{}

	for _, r2Obj := range r2ObjectList {
		metaResp, err := s.client.HeadObject(s.ctx, &r2.HeadObjectInput{
			Bucket: &s.bucket,
			Key:    aws.String(r2Obj.Key),
		})
		if err != nil {
			return nil, fmt.Errorf("fail to get metadata for object %s: %w", r2Obj.Key, err)
		}

		metaDataMap = append(metaDataMap, metaResp.Metadata)
	}

	for k, v := range metaDataMap {
		r2ObjectList[k].Metadata = v["uploadedby"]
	}

	sorted := sortStructByTime(r2ObjectList)

	return sorted, nil
}

// createR2Client creates a new R2 client (unexported, used only by NewR2Service)
func createR2Client(accessKeyID string, accessKeySecret string, accountID string) (*r2.Client, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("fail to load AWS config: %w", err)
	}

	client := r2.NewFromConfig(cfg, func(o *r2.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})

	return client, nil
}
