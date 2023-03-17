package main

import (
	"container_provisioner/api"
)

func main() {

	// Start the API server
	api.ServerInit()

	// Load the API routes
	api.Router()

	// exportedFileName := containers.Provision("https://www.tripadvisor.com/Hotel_Review-g188107-d199124-Reviews-Hotel_Des_Voyageurs-Lausanne_Canton_of_Vaud.html")

	// // Read the creds from the JSON file
	// data := utils.ParseCredsFromJSON("./creds.json")

	// // Create a new R3 client
	// r2Client := utils.CreateR2Client(data.AccessKeyId, data.AccessKeySecret, data.AccountId)

	// // Read the exported csv file
	// file := utils.ReadFromFile(exportedFileName)

	// // Upload the file to R2
	// utils.R2UploadObject(r2Client, data.BucketName, exportedFileName, file)
}
