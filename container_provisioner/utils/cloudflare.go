package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	r2 "github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	// Read the creds from the JSON file
	data = ParseCredsFromJSON("./creds.json")
	// Create a new R2 client
	r2Client = CreateR2Client(data.AccessKeyId, data.AccessKeySecret, data.AccountId)
)

// R2 object struct
type R2Obj struct {
	ChecksumAlgorithm string `json:"checksumAlgorithm"`
	Etag              string `json:"Etag"`
	Key               string `json:"Key"`
	LastModified      string `json:"LastModified"`
	Size              int64  `json:"Size"`
	StorageClass      string `json:"StorageClass"`
	Owner             string `json:"Owner"`
}

// R2UploadObject upload an object to R2
func R2UploadObject(fileName string, fileData io.Reader) {

	// Upload an object to R2
	_, err := r2Client.PutObject(context.TODO(), &r2.PutObjectInput{
		Bucket: &data.BucketName,
		Key:    aws.String(fileName),
		Body:   fileData,
	})
	ErrorHandler(err)

	err = os.Remove(fileName)
	ErrorHandler(err)
}

// R2ListObjects List objects in R2
func R2ListObjects() []string {

	// List objects in R2
	listObjectsOutput, err := r2Client.ListObjectsV2(context.TODO(), &r2.ListObjectsV2Input{
		Bucket: &data.BucketName,
	})
	ErrorHandler(err)

	// String slice to hold the file names
	fileNames := []string{}

	// _ reqiored to ignore the error
	for _, object := range listObjectsOutput.Contents {

		obj, err := json.MarshalIndent(object, "", "\t")
		ErrorHandler(err)

		r2Obj := R2Obj{}

		err = json.Unmarshal([]byte(obj), &r2Obj)
		ErrorHandler(err)

		fileNames = append(fileNames, r2Obj.Key)
	}
	return fileNames
}

// CreateR2Client creates a new R2 client
func CreateR2Client(accessKeyId string, accessKeySecret string, accountId string) *r2.Client {

	// Logic from the documentation
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service string, region string, options ...interface{}) (aws.Endpoint, error) {

		// Logic from the documentation
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	// Load the default configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
	)
	ErrorHandler(err)

	client := r2.NewFromConfig(cfg)

	return client
}
