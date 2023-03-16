package main

import (
	"container_provisioner/api"
	"container_provisioner/containers"
)

func main() {
	containers.Provision()
	api.PrintText()
}
