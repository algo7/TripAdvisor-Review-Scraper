package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	bucketName = "reviews"
	s3Client   = createS3Client()
)

// R2UploadObject upload an object to R2
func R2UploadObject(fileName string, fileData io.Reader) {

	// Upload an object to R2
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    aws.String(fileName),
		Body:   fileData,
	})
	ErrorHandler(err)
}

// r2ListObjects List objects in R2
func r2ListObjects() {

	// List objects in R2
	listObjectsOutput, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	ErrorHandler(err)

	for _, object := range listObjectsOutput.Contents {
		obj, _ := json.MarshalIndent(object, "", "\t")
		fmt.Println(string(obj))
	}
}

// Create a new S3 client
func createS3Client(accessKeyId string, accessKeySecret string, accountId string) *s3.Client {

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

	client := s3.NewFromConfig(cfg)

	return client
}
