package api

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/views"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v3"
)

const appTitle = "Algo7 TripAdvisor Scraper"

// ServerInstantiate is a function to instantiate the web server
func ServerInstantiate() *fiber.App {

	// The templates are compiled into the binary, see the views package
	engine := html.NewFileSystem(http.FS(views.Files), ".html")

	// Layout defines the variable name that is used to yield templates within layouts
	engine.Layout("embed") // Optional. Default: "embed"

	// Delims sets the action delimiters to the specified strings
	engine.Delims("{{", "}}") // Optional. Default: engine delimiters
	engine.Reload(false)
	// Custom config
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Algo7 TripAdvisor Scraper",
		AppName:       "Algo7 TripAdvisor Scraper v1.0.0",
		Views:         engine,
		// Every page is rendered inside the shared layout
		ViewsLayout: "layouts/main",
		// Errors are shown as an HTML page instead of plain text or JSON
		ErrorHandler: errorPage,
	})

	// Healthcheck /livez and /readyz routes
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New())

	// Stylesheets, compressed and cached by the browser for an hour
	staticFiles, err := fs.Sub(views.Files, "static")
	if err != nil {
		log.Fatalf("fail to load the static files: %v", err)
	}
	app.Use("/static", static.New("", static.Config{
		FS:       staticFiles,
		Compress: true,
		MaxAge:   int(time.Hour.Seconds()),
	}))

	return app
}

// errorPage renders an error as an HTML page. The message of a fiber.Error is written for users and shown as is.
// Any other error is logged and replaced with a general message.
func errorPage(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	heading := "Something went wrong"
	message := "Please try again in a moment. If it keeps happening, email support."

	fiberErr, ok := errors.AsType[*fiber.Error](err)
	if ok {
		code = fiberErr.Code
		message = fiberErr.Message
	} else {
		log.Printf("%s %s failed: %v", c.Method(), c.Path(), err)
	}

	if code == fiber.StatusNotFound {
		heading = "Page not found"
		message = "This page doesn't exist."
	}

	renderErr := c.Status(code).Render("error", fiber.Map{
		"Title":   appTitle,
		"Heading": heading,
		"Message": message,
	})
	if renderErr != nil {
		log.Printf("fail to render the error page: %v", renderErr)
		return c.Status(code).SendString(message)
	}

	return nil
}
