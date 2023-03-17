package main

import (
	"container_provisioner/containers"
	"container_provisioner/utils"
)

func main() {
	containers.Provision()

	// Read the creds from the JSON file
	data := utils.ParseCredsFromJSON("./creds.json")

	// Create a new S3 client
	s3Client := utils.CreateS3Client(data.AccessKeyId, data.AccessKeySecret, data.AccountId)

	// Read the exported csv file
	file := utils.ReadFromFile("Reviews.csv")

	utils.R2UploadObject(s3Client, "test.txt", file)

}
