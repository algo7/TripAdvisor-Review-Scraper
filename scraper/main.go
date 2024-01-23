package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/tripadvisor"
)

func main() {
	// Scraper variables
	queryType := os.Getenv("QUERY_TYPE")
	parsedLocationID, err := strconv.Atoi(os.Getenv("LOCATION_ID"))
	if err != nil {
		log.Fatal("The location ID must be an positive integer")
	}
	locationID := uint32(parsedLocationID)

	// Get the query ID for the given query type.
	queryID := tripadvisor.GetQueryID(queryType)
	if err != nil {
		log.Fatal("The location ID must be an positive integer")
	}

	// Get the proxy host if set
	proxyHost := os.Getenv("PROXY_HOST")

	client := &http.Client{}

	// If the proxy host is set, use the proxy client
	if proxyHost != "" {
		client, err = tripadvisor.GetHTTPClientWithProxy(proxyHost)
		if err != nil {
			log.Fatalf("Error creating HTTP client with the give proxy %s: %v", proxyHost, err)
		}
		log.Printf("Using proxy: %s", proxyHost)
		// Check IP
		ip, err := tripadvisor.CheckIP(client)
		if err != nil {
			log.Fatalf("Error checking IP: %v", err)
		}
		log.Printf("IP: %s", ip)
	}

	// CSV file
	fileName := "reviews.csv"
	headers := []string{"Title", "Text", "Rating", "Year", "Month", "Day"}

	// Fetch the review count for the given location ID
	reviewCount, err := tripadvisor.FetchReviewCount(client, locationID, queryType)
	if err != nil {
		log.Fatalf("Error fetching review count: %v", err)
	}
	if reviewCount == 0 {
		log.Fatalf("No reviews found for location ID %d", locationID)
	}
	log.Printf("Review count: %d", reviewCount)

	// Create a file to save the CSV data
	fileHandle, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error creating file %s: %v", fileName, err)
	}
	defer fileHandle.Close()

	// Create a new csv writer
	writer := csv.NewWriter(fileHandle)
	defer writer.Flush()

	// Writing header to the CSV file
	err = writer.Write(headers)
	if err != nil {
		log.Fatalf("Error writing header to csv: %v", err)
	}

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
		resp, err := tripadvisor.MakeRequest(client, queryID, "en", locationID, offset, 20)
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

			// Iterating over the reviews
			for _, row := range reviews {
				row := []string{
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

	// Write data to the CSV file
	err = writer.WriteAll(dataToWrite)
	if err != nil {
		log.Fatalf("Error writing data to csv: %v", err)
	}
}

func init() {
	// Check if the environment variables are set
	if os.Getenv("QUERY_TYPE") == "" {
		log.Fatal("QUERY_TYPE not set")
	}
	if os.Getenv("LOCATION_ID") == "" {
		log.Fatal("LOCATION_ID not set")
	}
}
