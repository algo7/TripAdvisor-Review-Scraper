package main

import (
	"container_provisioner/api"
	"container_provisioner/database"
)

func main() {

	// Check if the redis server is up and running
	database.RedisConnectionCheck()

	// Load the API routes
	api.Router()

}
