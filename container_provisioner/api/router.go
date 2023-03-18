package api

import (
	"container_provisioner/utils"

	"github.com/gofiber/fiber/v2"
)

// Router is the main router for the API
func Router() {

	app := ServerInstantiate()

	app.Get("/", getMain)
	app.Post("/submit", postProvision)
	app.Get("/logs/:id", getLogs)
	app.Get("/logs-viewer", func(c *fiber.Ctx) error {
		return c.SendFile("./views/logs.html")
	})
	err := app.Listen(":3000")
	utils.ErrorHandler(err)
}
