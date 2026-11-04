package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/scrape"
	"github.com/gofiber/fiber/v3"
	"github.com/moby/moby/client"
)

// noContainers is a Docker client without any containers. Every other Docker call panics,
// because the embedded interface is nil.
type noContainers struct {
	containers.ContainerClient
}

func (noContainers) ContainerList(_ context.Context, _ client.ContainerListOptions) (client.ContainerListResult, error) {
	return client.ContainerListResult{}, nil
}

func newTestApp() *fiber.App {
	cm := &containers.ContainerManager{Client: noContainers{}}
	return Router(scrape.NewScraper(cm, nil, nil))
}

// send sends a request to the app and returns the status code, the content type and the body.
func send(t *testing.T, app *fiber.App, req *http.Request) (int, string, string) {
	t.Helper()
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, resp.Header.Get("Content-Type"), string(body)
}

func TestPagesRenderInsideTheLayout(t *testing.T) {
	app := newTestApp()

	tests := []struct {
		path     string
		wantText string
	}{
		{path: "/", wantText: "Start scraping"},
		{path: "/tasks", wantText: "No scrapes are running right now."},
		{path: "/logs-viewer?container_id=abc", wantText: "Scrape progress"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			status, contentType, body := send(t, app, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if status != http.StatusOK || !strings.HasPrefix(contentType, "text/html") {
				t.Fatalf("status %d, content type %q, want 200 text/html", status, contentType)
			}
			for _, want := range []string{tt.wantText, `href="/static/pico.classless.min.css"`, `name="viewport"`} {
				if !strings.Contains(body, want) {
					t.Errorf("page is missing %q", want)
				}
			}
		})
	}
}

func TestStaticFilesAreServed(t *testing.T) {
	app := newTestApp()

	for _, path := range []string{"/static/pico.classless.min.css", "/static/app.css"} {
		status, contentType, body := send(t, app, httptest.NewRequest(http.MethodGet, path, nil))
		if status != http.StatusOK || !strings.HasPrefix(contentType, "text/css") || body == "" {
			t.Errorf("%s: status %d, content type %q, %d bytes, want 200 text/css", path, status, contentType, len(body))
		}
	}
}

func TestUnknownPageShowsTheErrorPage(t *testing.T) {
	status, contentType, body := send(t, newTestApp(), httptest.NewRequest(http.MethodGet, "/nope", nil))

	if status != http.StatusNotFound || !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("status %d, content type %q, want 404 text/html", status, contentType)
	}
	if !strings.Contains(body, "Page not found") || !strings.Contains(body, `href="/static/pico.classless.min.css"`) {
		t.Errorf("404 page isn't the error page inside the layout:\n%s", body)
	}
}

func TestSubmitWithAnInvalidURLExplainsTheProblem(t *testing.T) {
	form := url.Values{
		"url":               {"https://www.tripadvisor.fr/Hotel_Review-g1-d2-Reviews-X-Y.html"},
		"upload_identifier": {"test"},
		"scrape_option":     {"HOTEL"},
	}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	status, _, body := send(t, newTestApp(), req)

	if status != http.StatusOK || !strings.Contains(body, "The scrape didn't start") || !strings.Contains(body, "Invalid HOTEL URL") {
		t.Errorf("status %d, want the submission error page with the reason:\n%s", status, body)
	}
}

func TestLogsOfAFinishedScrapeExplainWhereTheFileIs(t *testing.T) {
	status, contentType, body := send(t, newTestApp(), httptest.NewRequest(http.MethodGet, "/logs/abc", nil))

	if status != http.StatusBadRequest || !strings.HasPrefix(contentType, "application/json") || !strings.Contains(body, "Downloads page") {
		t.Errorf("status %d, content type %q, body %s, want a 400 JSON error pointing to Downloads", status, contentType, body)
	}
}
