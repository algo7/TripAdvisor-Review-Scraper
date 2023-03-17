package api

import (
	"github.com/gofiber/fiber/v2"
)

var app = App

func Router() {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// app.Post("/submit", scrape)
}
