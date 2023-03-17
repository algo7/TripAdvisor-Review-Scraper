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
	Metadata          string
}

// R2UploadObject upload an object to R2
func R2UploadObject(fileName string, uploadIdentifier string, fileData io.Reader) {

	// Upload an object to R2
	_, err := r2Client.PutObject(context.TODO(), &r2.PutObjectInput{
		Bucket: &data.BucketName,
		Key:    aws.String(fileName),
		Body:   fileData,
		Metadata: map[string]string{
			"uploadedby": uploadIdentifier,
		},
	})
	ErrorHandler(err)

	err = os.Remove(fileName)
	ErrorHandler(err)
}

// R2ListObjects List objects in R2 and return a string slice of the file names
func R2ListObjects() []R2Obj {

	// List objects in R2
	listObjectsOutput, err := r2Client.ListObjectsV2(context.TODO(), &r2.ListObjectsV2Input{
		Bucket:     &data.BucketName,
		FetchOwner: true,
	})
	ErrorHandler(err)

	// String slice to hold the file names
	files := []R2Obj{}

	// _ required to ignore the error
	for _, object := range listObjectsOutput.Contents {

		// Marshal the object to JSON in a pretty format
		obj, err := json.MarshalIndent(object, "", "\t")
		ErrorHandler(err)

		// Create a new R2 object
		r2Obj := R2Obj{}

		// Unmarshal the JSON into a struct
		err = json.Unmarshal([]byte(obj), &r2Obj)
		ErrorHandler(err)

		// Call HeadObject to retrieve metadata for the object
		metaResp, err := r2Client.HeadObject(context.TODO(), &r2.HeadObjectInput{
			Bucket: &data.BucketName,
			Key:    aws.String(r2Obj.Key),
		})

		// Enrich the full R2 object with the metadata
		for k, v := range metaResp.Metadata {
			fmt.Println(k, v)
			// Check if the key is UploadedBy
			// Map keys are turned into lowercase by the SDK
			if k == "uploadedby" {
				r2Obj.Metadata = v
			}
		}

		// Append the object to the files slice
		files = append(files, r2Obj)
	}

	return files
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
