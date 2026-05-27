package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/internal/config"
	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/tripadvisor"
	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/utils"
)

func main() {
	// Scraper variables
	var allReviews []tripadvisor.Review
	var michelinInfo *tripadvisor.MichelinInfo

	config, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error creating scrape config: %v", err)
	}

	// Get the query type from the URL
	queryType := tripadvisor.GetURLType(config.LocationURL)
	if queryType == "" {
		log.Fatal("Invalid URL")
	}
	log.Printf("Location Type: %s", queryType)

	// Parse the location ID and location name from the URL
	locationID, geoID, locationName, err := tripadvisor.ParseURL(config.LocationURL, queryType)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}
	log.Printf("Location ID: %d", locationID)
	log.Printf("Location Name: %s", locationName)

	// Get the query ID for the given query type.
	queryID := tripadvisor.GetQueryID(queryType)

	// The default HTTP client
	client := &http.Client{
		Transport: http.DefaultTransport,
	}

	// If the proxy host is set, use the proxy client
	if config.ProxyHost != "" {

		// Get the HTTP client with the proxy
		client, err = tripadvisor.GetHTTPClientWithProxy(config.ProxyHost)
		if err != nil {
			log.Fatalf("Error creating HTTP client with the give proxy %s: %v", config.ProxyHost, err)
		}

		// Check IP
		ip, err := utils.CheckIP(client)
		if err != nil {
			log.Fatalf("Error checking IP: %v", err)
		}
		log.Printf("Proxy IP: %s", ip)
	}

	// Fetch the review count for the given location ID
	reviewCount, err := tripadvisor.FetchReviewCount(client, locationID, geoID, queryType, config.Languages)
	if err != nil {
		log.Fatalf("Error fetching review count: %v", err)
	}
	if reviewCount == 0 {
		log.Fatalf("No reviews found for location ID %d", locationID)
	}
	log.Printf("Review count: %d", reviewCount)

	// Create a file to save the reviews data
	fileName := fmt.Sprintf("reviews.%s", config.FileType)
	fileHandle, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error creating file %s: %v", fileName, err)
	}
	defer fileHandle.Close()

	// Calculate the number of iterations required to fetch all reviews
	iterations := tripadvisor.CalculateIterations(uint32(reviewCount))
	log.Printf("Total Iterations: %d", iterations)

	// Create a slice to store the data to be written to the CSV file
	dataToWrite := make([][]string, 0, reviewCount)

	// Scrape the reviews
	for i := range iterations {

		// Introduce random delay to avoid getting blocked. The delay is between 1 and 5 seconds
		delay := rand.Intn(5) + 1
		log.Printf("Iteration: %d. Delaying for %d seconds", i, delay)
		time.Sleep(time.Duration(delay) * time.Second)

		// Calculate the offset for the current iteration
		offset := tripadvisor.CalculateOffset(i)

		// Make the request to the TripAdvisor GraphQL endpoint
		resp, err := tripadvisor.MakeRequest(client, queryID, queryType, config.Languages, locationID, geoID, offset, 20)
		if err != nil {
			log.Fatalf("Error making request at iteration %d: %v", i, err)
		}

		// Extract reviews using the shared helper (handles both ReviewsProxy and Locations paths)
		reviews := tripadvisor.ExtractReviews(resp)

		// Extract Michelin info once from the first response that contains it
		if michelinInfo == nil {
			michelinInfo = tripadvisor.ExtractMichelinInfo(resp)
		}

		// Append the reviews to the allReviews slice
		allReviews = append(allReviews, reviews...)

		if config.FileType == "csv" {
			for _, r := range reviews {
				dataToWrite = append(dataToWrite, tripadvisor.ReviewToCSVRow(r, locationName, michelinInfo))
			}
		}
	}

	if config.FileType == "csv" {
		writer := csv.NewWriter(fileHandle)

		// Write CSV headers (includes Michelin columns when Michelin data is present)
		if err := writer.Write(tripadvisor.CSVHeaders(michelinInfo != nil)); err != nil {
			log.Fatalf("Error writing header to csv: %v", err)
		}

		// Write all review rows
		if err := writer.WriteAll(dataToWrite); err != nil {
			log.Fatalf("Error writing data to csv: %v", err)
		}
	}

	// If the file type is JSON, write the complete scrape result (reviews + Michelin data)
	if config.FileType == "json" {
		tripadvisor.SortReviewsByDate(allReviews)
		result := &tripadvisor.ScrapeResult{
			Reviews:  allReviews,
			Michelin: michelinInfo,
		}
		if err := tripadvisor.WriteScrapeResultToJSONFile(result, fileHandle); err != nil {
			log.Fatalf("Error writing data to JSON file: %v", err)
		}
	}

	log.Printf("Data written to %s", fileName)
	log.Println("Scraping completed")
}
