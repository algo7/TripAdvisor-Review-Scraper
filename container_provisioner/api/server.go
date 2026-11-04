package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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

	// The stylesheets are compiled into the binary, see the views package
	staticFiles, err := fs.Sub(views.Files, "static")
	if err != nil {
		log.Fatalf("fail to load the static files: %v", err)
	}
	urls, err := staticURLs(staticFiles)
	if err != nil {
		log.Fatalf("fail to hash the static files: %v", err)
	}

	// The templates are compiled into the binary, see the views package
	engine := html.NewFileSystem(http.FS(views.Files), ".html")

	// {{static "app.css"}} returns the file's URL with a hash of its content
	engine.AddFunc("static", func(name string) (string, error) {
		url, ok := urls[name]
		if !ok {
			return "", fmt.Errorf("unknown static file %q", name)
		}
		return url, nil
	})

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

	// Stylesheets, compressed. Pages link them with a content hash in the URL, so browsers can keep them
	// for a year: a changed file gets a new URL.
	app.Use("/static", static.New("", static.Config{
		FS:       staticFiles,
		Compress: true,
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
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

// staticURLs returns the URL of every static file with a hash of its content, e.g. /static/app.css?v=3f9c1a2b4d5e.
// Browsers cache files by URL, so a changed file is loaded again instead of an older cached copy being used.
func staticURLs(files fs.FS) (map[string]string, error) {
	urls := map[string]string{}

	err := fs.WalkDir(files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		content, err := fs.ReadFile(files, path)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(content)
		urls[path] = "/static/" + path + "?v=" + hex.EncodeToString(hash[:])[:12]

		return nil
	})

	return urls, err
}
