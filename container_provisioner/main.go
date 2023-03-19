package main

import (
	"container_provisioner/api"
	"container_provisioner/database"
)

func main() {

	// Load the API routes
	api.Router()

	database.RedisConnectionCheck()
}
