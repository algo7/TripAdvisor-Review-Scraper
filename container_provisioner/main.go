package main

import (
	"container_provisioner/containers"
	"container_provisioner/utils"
)

func main() {
	exportedFileName := containers.Provision()

	// Read the creds from the JSON file
	data := utils.ParseCredsFromJSON("./creds.json")

	// Create a new S3 client
	s3Client := utils.CreateS3Client(data.AccessKeyId, data.AccessKeySecret, data.AccountId)

	// Read the exported csv file
	file := utils.ReadFromFile(exportedFileName)

	// Upload the file to R2
	utils.R2UploadObject(s3Client, data.BucketName, exportedFileName, file)
}
