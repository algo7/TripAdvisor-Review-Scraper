package api

import (
	"container_provisioner/utils"
)

// Router is the main router for the API
func Router() {

	app := ServerInstantiate()

	app.Get("/", getMain)
	app.Post("/submit", postProvision)
	app.Get("/logs/:id", getLogs)
	app.Get("/logs-viewer", getLogsViewer)
	app.Get("/tasks", getRunningTasks)

	err := app.Listen(":3000")
	utils.ErrorHandler(err)
}
