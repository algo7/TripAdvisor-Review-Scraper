package main

import (
	"log"

	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/tripadvisor"
)

func main() {

	count, err := tripadvisor.FetchReviewCount(231860, "HOTEL")
	if err != nil {
		log.Fatalf("Error fetching review count: %v", err)
	}
	log.Printf("Review count: %d", count)

}
