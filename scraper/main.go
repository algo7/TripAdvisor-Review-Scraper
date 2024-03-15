package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/internal/config"
	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/tripadvisor"
	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/utils"
)

func main() {
	// Scraper variables
	var allReviews []tripadvisor.Review
	var location tripadvisor.Location

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
	locationID, locationName, err := tripadvisor.ParseURL(config.LocationURL, queryType)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}
	log.Printf("Location ID: %d", locationID)
	log.Printf("Location Name: %s", locationName)

	// Get the query ID for the given query type.
	queryID := tripadvisor.GetQueryID(queryType)
	if err != nil {
		log.Fatal("The location ID must be a positive integer")
	}

	// The default HTTP client
	client := &http.Client{}

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
	reviewCount, err := tripadvisor.FetchReviewCount(client, locationID, queryType, config.Languages)
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
	for i := uint32(0); i < iterations; i++ {

		// Introduce random delay to avoid getting blocked. The delay is between 1 and 5 seconds
		delay := rand.Intn(5) + 1
		log.Printf("Iteration: %d. Delaying for %d seconds", i, delay)
		time.Sleep(time.Duration(delay) * time.Second)

		// Calculate the offset for the current iteration
		offset := tripadvisor.CalculateOffset(i)

		// Make the request to the TripAdvisor GraphQL endpoint
		resp, err := tripadvisor.MakeRequest(client, queryID, config.Languages, locationID, offset, 20)
		if err != nil {
			log.Fatalf("Error making request at iteration %d: %v", i, err)
		}

		// Check if responses is nil before dereferencing
		if resp == nil {
			log.Fatalf("Received nil response for location ID %d at iteration: %d", locationID, i)
		}

		// Now it's safe to dereference responses
		response := *resp

		// Check if the response is not empty and if the response contains reviews
		if len(response) > 0 && len(response[0].Data.Locations) > 0 {

			// Get the reviews from the response
			reviews := response[0].Data.Locations[0].ReviewListPage.Reviews

			// Append the reviews to the allReviews slice
			allReviews = append(allReviews, reviews...)

			// Store the location data
			location = response[0].Data.Locations[0].Location

			if config.FileType == "csv" {
				// Iterating over the reviews
				for _, row := range reviews {
					row := []string{
						locationName,
						row.Title,
						row.Text,
						strconv.Itoa(row.Rating),
						row.CreatedDate[0:4],
						row.CreatedDate[5:7],
						row.CreatedDate[8:10],
					}

					// Append the row to the dataToWrite slice
					dataToWrite = append(dataToWrite, row)
				}
			}

		}

	}

	if config.FileType == "csv" {
		// Create a new csv writer. We are using writeAll so defer writer.Flush() is not required
		writer := csv.NewWriter(fileHandle)

		// Writing header to the CSV file
		headers := []string{"Location Name", "Title", "Text", "Rating", "Year", "Month", "Day"}
		err = writer.Write(headers)
		if err != nil {
			log.Fatalf("Error writing header to csv: %v", err)
		}
		// Write data to the CSV file
		err = writer.WriteAll(dataToWrite)
		if err != nil {
			log.Fatalf("Error writing data to csv: %v", err)
		}
	}

	// If the file type is JSON, write the data to the file
	if config.FileType == "json" {
		// Sort the reviews by date
		tripadvisor.SortReviewsByDate(allReviews)

		// Write the data to the JSON file
		err := tripadvisor.WriteReviewsToJSONFile(allReviews, location, fileHandle)
		if err != nil {
			log.Fatalf("Error writing data to JSON file: %v", err)
		}
	}
	log.Printf("Data written to %s", fileName)
	log.Println("Scrapping completed")
}
