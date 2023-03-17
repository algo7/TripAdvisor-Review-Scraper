package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	r2 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2UploadObject upload an object to R2
func R2UploadObject(r2Client *r2.Client, bucketName string, fileName string, fileData io.Reader) {

	// Upload an object to R2
	_, err := r2Client.PutObject(context.TODO(), &r2.PutObjectInput{
		Bucket: &bucketName,
		Key:    aws.String(fileName),
		Body:   fileData,
	})
	ErrorHandler(err)
}

// r2ListObjects List objects in R2
func r2ListObjects(r2Client *r2.Client, bucketName string) {

	// List objects in R2
	listObjectsOutput, err := r2Client.ListObjectsV2(context.TODO(), &r2.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	ErrorHandler(err)

	for _, object := range listObjectsOutput.Contents {
		obj, _ := json.MarshalIndent(object, "", "\t")
		fmt.Println(string(obj))
	}
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
