package tripadvisor

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// WriteReviewToCSV is a function that writes data to a CSV file.
func WriteReviewToCSV(fileHandle *os.File, headers []string, reviews Reviews) error {

	// Create a new csv writer
	writer := csv.NewWriter(fileHandle)
	defer writer.Flush()

	// Writing header to the CSV file
	err := writer.Write(headers)
	if err != nil {
		return fmt.Errorf("Error writing header to csv: %v", err)
	}

	// Iterating over the reviews and writing to the CSV file
	for _, row := range reviews {
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
			return fmt.Errorf("Error writing row to csv: %v", err)
		}
	}

	return nil
}
