package main

import (
	"container_provisioner/containers"
	"container_provisioner/utils"
)

func main() {
	exportedFileName := containers.Provision()

	// Read the creds from the JSON file
	data := utils.ParseCredsFromJSON("./creds.json")

	// Create a new R3 client
	r2Client := utils.CreateR2Client(data.AccessKeyId, data.AccessKeySecret, data.AccountId)

	// Read the exported csv file
	file := utils.ReadFromFile(exportedFileName)

	// Upload the file to R2
	utils.R2UploadObject(r2Client, data.BucketName, exportedFileName, file)
}
