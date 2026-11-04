package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

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
			for _, want := range []string{tt.wantText, `href="/static/pico.classless.min.css?v=`, `name="viewport"`} {
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
	if !strings.Contains(body, "Page not found") || !strings.Contains(body, `href="/static/pico.classless.min.css?v=`) {
		t.Errorf("404 page isn't the error page inside the layout:\n%s", body)
	}
}

// submit posts the form and returns the status code and the body
func submit(t *testing.T, link string, uploadIdentifier string) (int, string) {
	t.Helper()
	form := url.Values{
		"url":               {link},
		"upload_identifier": {uploadIdentifier},
	}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	status, _, body := send(t, newTestApp(), req)
	return status, body
}

func TestFormHasNoTypeDropdown(t *testing.T) {
	_, _, body := send(t, newTestApp(), httptest.NewRequest(http.MethodGet, "/", nil))

	if strings.Contains(body, "<select") || strings.Contains(body, "scrape_option") {
		t.Error("the form still asks for the scrape type")
	}
}

func TestSubmitProblemsAreShownOnTheForm(t *testing.T) {
	tests := []struct {
		name             string
		link             string
		uploadIdentifier string
		wantStatus       int
		wantText         string
		wantInvalid      string
	}{
		{
			name:             "country site",
			link:             "https://www.tripadvisor.fr/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne.html",
			uploadIdentifier: "rocket_fox",
			wantStatus:       http.StatusUnprocessableEntity,
			wantText:         "Country sites like tripadvisor.fr",
			wantInvalid:      "url",
		},
		{
			name:             "not a supported page",
			link:             "https://www.tripadvisor.com/Search?q=lausanne",
			uploadIdentifier: "rocket_fox",
			wantStatus:       http.StatusUnprocessableEntity,
			wantText:         "isn&#39;t a TripAdvisor hotel, restaurant, attraction or airline page",
			wantInvalid:      "url",
		},
		{
			name:             "missing upload ID",
			link:             "https://www.tripadvisor.com/Airline_Review-d8729113-Reviews-Lufthansa",
			uploadIdentifier: "   ",
			wantStatus:       http.StatusUnprocessableEntity,
			wantText:         "Enter an upload ID",
			wantInvalid:      "upload_identifier",
		},
		{
			// The attraction link passes validation. The test app has no proxies, so the scrape doesn't start.
			name:             "attraction link is accepted",
			link:             "https://www.tripadvisor.com/Attraction_Review-g187261-d1008501-Reviews-Les_Ailes_du_Mont_Blanc-Chamonix.html#REVIEWS",
			uploadIdentifier: "rocket_fox",
			wantStatus:       http.StatusServiceUnavailable,
			wantText:         "All proxies are busy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := submit(t, tt.link, tt.uploadIdentifier)

			if status != tt.wantStatus {
				t.Errorf("status %d, want %d", status, tt.wantStatus)
			}
			for _, want := range []string{"The scrape didn't start.", tt.wantText, "Start scraping"} {
				if !strings.Contains(body, want) {
					t.Errorf("page is missing %q", want)
				}
			}
			// The typed upload ID is kept
			if strings.TrimSpace(tt.uploadIdentifier) != "" && !strings.Contains(body, `value="`+tt.uploadIdentifier+`"`) {
				t.Errorf("the upload ID %q wasn't kept in the form", tt.uploadIdentifier)
			}
			invalidInputs := strings.Count(body, `aria-invalid="true"`)
			if tt.wantInvalid == "" && invalidInputs != 0 {
				t.Errorf("%d fields are marked invalid, want none", invalidInputs)
			}
			if tt.wantInvalid != "" && (invalidInputs != 1 || !strings.Contains(body, `name="`+tt.wantInvalid+`"`)) {
				t.Errorf("%d fields are marked invalid, want only %s", invalidInputs, tt.wantInvalid)
			}
		})
	}
}

func TestLogsOfAFinishedScrapeExplainWhereTheFileIs(t *testing.T) {
	status, contentType, body := send(t, newTestApp(), httptest.NewRequest(http.MethodGet, "/logs/abc", nil))

	if status != http.StatusBadRequest || !strings.HasPrefix(contentType, "application/json") || !strings.Contains(body, "Downloads page") {
		t.Errorf("status %d, content type %q, body %s, want a 400 JSON error pointing to Downloads", status, contentType, body)
	}
}

func TestStaticURLsChangeWithTheContent(t *testing.T) {
	before, err := staticURLs(fstest.MapFS{"app.css": {Data: []byte("a { color: red; }")}})
	if err != nil {
		t.Fatal(err)
	}
	after, err := staticURLs(fstest.MapFS{"app.css": {Data: []byte("a { color: blue; }")}})
	if err != nil {
		t.Fatal(err)
	}
	again, err := staticURLs(fstest.MapFS{"app.css": {Data: []byte("a { color: blue; }")}})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(after["app.css"], "/static/app.css?v=") {
		t.Fatalf("URL %q, want /static/app.css?v=<hash>", after["app.css"])
	}
	if before["app.css"] == after["app.css"] {
		t.Error("the URL stayed the same after the file changed, browsers would keep the old copy")
	}
	if after["app.css"] != again["app.css"] {
		t.Error("the URL changed although the file didn't, browsers would load it again for nothing")
	}
}
