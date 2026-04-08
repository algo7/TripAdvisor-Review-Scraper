package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Compile the regular expression and store it in a package-level variable
var (
	tripAdvisorHotelURLRegexp   = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Hotel_Review-g\d{6,10}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`)
	tripAdvisorRestaurantRegexp = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Restaurant_Review-g\d{6,10}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`)
	tripAdvisorAirlineRegexp    = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Airline_Review-d\d{6,10}-Reviews-[\w-]{1,255}$`)
)

// WriteToFileFromTarStream writes a file to disk
func WriteToFileFromTarStream(fileName string, fileSuffix string, tarF io.ReadCloser) (string, error) {

	// Untar the file
	// Note: This is not a generic untar function. It only works for a single file
	/**
		A tar file is a collection of binary data segments (usually sourced from files). Each segment starts with a header that contains metadata about the binary data, that follows it, and how to reconstruct it as a file.

	+---------------------------+
	| [name][mode][uid][guild]  |
	| ...                       |
	+---------------------------+
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	+---------------------------+
	| [name][mode][uid][guild]  |
	| ...                       |
	+---------------------------+
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	+---------------------------+
		**/

	// Read the tar file
	tarReader := tar.NewReader(tarF)

	// Get the tar header and go to the next entry in the tar file
	_, err := tarReader.Next()
	if err != nil {
		return "", fmt.Errorf("fail to read the tar file: %w", err)
	}

	fileNameToWrite := fileName + "-" + fileSuffix + ".csv"

	// Create the file
	out, err := os.Create(fileNameToWrite)
	if err != nil {
		return "", fmt.Errorf("fail to create file to hold the extracted data: %w", err)
	}
	defer out.Close()

	// Write the file to disk
	_, err = io.Copy(out, tarReader)
	if err != nil {
		return "", fmt.Errorf("fail to write the extracted file to disk: %w", err)
	}

	// Return the file name
	return fileNameToWrite, nil
}

// ReadFromFile reads a file from disk
func ReadFromFile(fileName string) (*os.File, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("fail to read extracted csv file: %w", err)
	}

	return file, nil
}

// GetLocationNameFromURL get the scrape target name from the given URL
func GetLocationNameFromURL(url string, scrapOption string) string {

	// Split the url by "-"
	splitURL := strings.Split(url, "-")

	switch scrapOption {
	case "HOTEL", "RESTO":
		return splitURL[4]
	case "AIRLINE":
		return strings.Join(splitURL[3:], "_")
	default:
		return ""
	}
}

// ValidateTripAdvisorURL validates the TripAdvisor URLs
func ValidateTripAdvisorURL(url string, scrapOption string) bool {
	switch scrapOption {
	case "HOTEL":
		return tripAdvisorHotelURLRegexp.MatchString(url)
	case "RESTO":
		return tripAdvisorRestaurantRegexp.MatchString(url)
	case "AIRLINE":
		return tripAdvisorAirlineRegexp.MatchString(url)
	default:
		return false
	}
}

// ValidateEmailAddress validates the EHL email address
func ValidateEmailAddress(email string) bool {
	regex := `^[a-z]+(\.[a-z]+)*@ehl\.ch$`
	match, _ := regexp.MatchString(regex, email)
	return match
}

// GenerateUUID generates a UUID
func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()[:11]
}

// ParseTime converts ISO 8601 time to a more readable format
func ParseTime(timeToParse string) (string, error) {
	// Parse the time string
	t, err := time.Parse(time.RFC3339Nano, timeToParse)
	if err != nil {
		return "", fmt.Errorf("fail to parse time string: %w", err)
	}

	// Format the time string in a more readable way
	formattedTime := t.Format("01/02/2006 15:04:05 MST")

	return formattedTime, nil
}
