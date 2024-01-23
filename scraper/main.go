package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/tripadvisor"
)

func main() {
	queryType := "HOTEL"
	locationID := uint32(231860)

	// Get the query ID for the given query type.
	queryID := tripadvisor.GetQueryID(queryType)

	count, err := tripadvisor.FetchReviewCount(locationID, queryType)
	if err != nil {
		log.Fatalf("Error fetching review count: %v", err)
	}
	log.Printf("Review count: %d", count)

	// Create a file to save the CSV data
	fileName := "reviews.csv"

	// Create a file to save the CSV data
	fileHandle, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error creating file %s: %v", fileName, err)
	}
	defer fileHandle.Close()

	// Write the reviews to the CSV file
	headers := []string{"Title", "Text", "Rating", "Year", "Month", "Day"}

	// Create a new csv writer
	writer := csv.NewWriter(fileHandle)
	defer writer.Flush()

	// Writing header to the CSV file
	err = writer.Write(headers)
	if err != nil {
		log.Fatalf("Error writing header to csv: %v", err)
	}

	// Calculate the number of iterations required to fetch all reviews
	iterations := tripadvisor.CalculateIterations(uint32(count))
	log.Printf("Iterations: %d", iterations)

	for i := uint32(0); i < iterations; i++ {

		// Introduce random delay to avoid getting blocked. The delay is between 1 and 5 seconds.
		delay := rand.Intn(5) + 1
		log.Printf("Delaying for %d seconds", delay)
		time.Sleep(time.Duration(delay) * time.Second)

		// Calculate the offset for the current iteration
		offset := tripadvisor.CalculateOffset(i)

		resp, err := tripadvisor.MakeRequest(queryID, "en", locationID, offset, 20)
		if err != nil {
			log.Fatalf("Error making request at iteration %d: %v", i, err)
		}

		// Check if responses is nil before dereferencing
		if resp == nil {
			log.Fatalf("received nil response for location ID %d", locationID)
		}

		// Now it's safe to dereference responses
		response := *resp
		if len(response) > 0 && len(response[0].Data.Locations) > 0 {

			// Iterating over the reviews and writing to the CSV file
			for _, row := range response[0].Data.Locations[0].ReviewListPage.Reviews {
				row := []string{
					row.Title,
					row.Text,
					strconv.Itoa(row.Rating),
					row.CreatedDate[0:4],
					row.CreatedDate[5:7],
					row.CreatedDate[8:10],
				}

				err := writer.Write(row)
				if err != nil {
					log.Fatalf("Error writing row to csv: %v", err)
				}
			}
		}

	}
}
