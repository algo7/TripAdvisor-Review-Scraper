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

var (
	// Read the creds from the JSON file
	data = ParseCredsFromJSON("./credentials/creds.json")
	// Create a new R2 client
	r2Client = CreateR2Client(data.AccessKeyID, data.AccessKeySecret, data.AccountID)
	ctx      = context.TODO()
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

// R2UploadObject upload an object to R2
func R2UploadObject(fileName string, uploadIdentifier string, fileData io.Reader) {

	// Upload an object to R2
	_, err := r2Client.PutObject(ctx, &r2.PutObjectInput{
		Bucket: &data.BucketName,
		Key:    aws.String(fileName),
		Body:   fileData,
		Metadata: map[string]string{
			"uploadedby": uploadIdentifier,
		},
	})
	ErrorHandler(err)

	log.Printf("File: %s uploaded", fileName)

	// Remove the file from the local filesystem
	err = os.Remove(fileName)
	ErrorHandler(err)
}

// R2ListObjects List objects in R2 and return a string slice of the file names
func R2ListObjects() []R2Obj {

	// List objects in R2
	listObjectsOutput, err := r2Client.ListObjectsV2(ctx, &r2.ListObjectsV2Input{
		Bucket:     &data.BucketName,
		FetchOwner: aws.Bool(true),
	})
	ErrorHandler(err)

	// String slice to hold the r2 object information
	files := []R2Obj{}

	// The logic below maps the JSON to the R2Obj struct and then appends the struct to the slice of the same type
	// _ required to ignore the error
	for _, object := range listObjectsOutput.Contents {

		// Marshal the object to JSON in a pretty format
		obj, err := json.MarshalIndent(object, "", "\t")
		ErrorHandler(err)

		// Create a new R2 object
		r2Obj := R2Obj{}

		// Unmarshal the JSON into the R2Obj struct
		err = json.Unmarshal([]byte(obj), &r2Obj)
		ErrorHandler(err)

		// Append the object to the files slice
		files = append(files, r2Obj)
	}

	return files
}

// CreateR2Client creates a new R2 client
func CreateR2Client(accessKeyID string, accessKeySecret string, accountID string) *r2.Client {

	// Logic from the documentation
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service string, region string, options ...interface{}) (aws.Endpoint, error) {

		// Logic from the documentation
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		}, nil
	})

	// Load the default configuration
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	ErrorHandler(err)

	client := r2.NewFromConfig(cfg)

	return client
}

// R2EnrichMetaData enriches the R2 object in the list with the metadata
func R2EnrichMetaData(r2ObjectList []R2Obj) []R2Obj {

	// A slice of map of string key-value pairs to hold the metadata
	metaDataMap := []map[string]string{}

	// Loop through the the R2 object list and get the metadata for each object
	for _, r2Obj := range r2ObjectList {

		// Call HeadObject to retrieve metadata for the object
		metaResp, err := r2Client.HeadObject(ctx, &r2.HeadObjectInput{
			Bucket: &data.BucketName,
			Key:    aws.String(r2Obj.Key),
		})
		ErrorHandler(err)

		// Append the metadata to the metaDataMap slice
		metaDataMap = append(metaDataMap, metaResp.Metadata)
	}

	// Enrich the R2 object list with the metadata
	for k, v := range metaDataMap {
		r2ObjectList[k].Metadata = v["uploadedby"]
	}

	sorted := sortStructByTime(r2ObjectList)

	return sorted
}
