package files

import (
	"fmt"
	"os"
)

// CreateReviewsCSV is a function that creates a CSV file of reviews.
func CreateReviewsCSV(fileName string) error {

	// Create a file to save the CSV data
	file, err := os.Create("reviews.csv")
	if err != nil {
		return fmt.Errorf("Error creating file %s: %v", fileName, err)
	}

	// Defer closing the file until the function returns
	defer file.Close()

	return nil
}
