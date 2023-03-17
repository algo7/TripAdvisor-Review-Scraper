package main

import "container_provisioner/utils"

func main() {
	// containers.Provision()

	utils.ParseCredsFromJSON("./creds.json")
}
