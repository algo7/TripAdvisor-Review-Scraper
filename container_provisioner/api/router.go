package api

import (
	"container_provisioner/utils"
	"fmt"
)

func Router() {

	app := ServerInstantiate()

	app.Get("/", mainView)

	fmt.Println("Server started on port 3000")
	err := app.Listen(":3000")
	utils.ErrorHandler(err)

}
