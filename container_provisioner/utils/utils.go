package utils

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Compile the regular expression and store it in a package-level variable
var (
	tripAdvisorHotelURLRegexp   = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Hotel_Review-g\d{6,10}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`)
	tripAdvisorRestaurantRegexp = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Restaurant_Review-g\d{6,10}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`)
	tripAdvisorAirlineRegexp    = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Airline_Review-d\d{6,10}-Reviews-[\w-]{1,255}\$`)
)

// Creds is the Credentials of the R2 bucket
type Creds struct {
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	AccountID       string `json:"accountId"`
	BucketName      string `json:"bucketName"`
}

// ErrorHandler is a generic error handler
func ErrorHandler(err error) {
	if err != nil {
		formattedError := fmt.Errorf("Error: %w", err)
		log.Fatalln(formattedError)
	}
}

// WriteToFileFromTarStream writes a file to disk
func WriteToFileFromTarStream(fileName string, fileSuffix string, tarF io.ReadCloser) string {

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
	ErrorHandler(err)

	fileNameToWrite := fileName + "-" + fileSuffix + ".csv"

	// Create the file
	out, err := os.Create(fileNameToWrite)
	ErrorHandler(err)
	defer out.Close()

	// Write the file to disk
	_, err = io.Copy(out, tarReader)
	ErrorHandler(err)

	// Return the file name
	return fileNameToWrite
}

// ReadFromFile reads a file from disk
func ReadFromFile(fileName string) *os.File {
	file, err := os.Open(fileName)

	ErrorHandler(err)

	return file
}

// ParseCredsFromJSON parses the credentials from a JSON file
func ParseCredsFromJSON(fileName string) Creds {
	// Read file
	file := ReadFromFile(fileName)
	defer file.Close()

	// Parse the JSON file
	decoder := json.NewDecoder(file)
	var creds Creds
	err := decoder.Decode(&creds)
	ErrorHandler(err)

	return creds
}

// GetScrapeTargetNameFromURL get the scrape target name from the given URL
func GetScrapeTargetNameFromURL(url string, scrapOption string) string {
	// Split the url by "-"
	splitURL := strings.Split(url, "-")

	switch scrapOption {
	case "HOTEL", "RESTO":
		return splitURL[4]
	case "AIRLINE":
		if len(splitURL) > 4 {
			return fmt.Sprintf("%s-%s", splitURL[4], splitURL[5])
		}
		return splitURL[3]
	default:
		return ""
	}
}

// ValidateTripAdvisorURL validates the TripAdvisor URLs
func ValidateTripAdvisorURL(url string, scrapOption string) bool {
	switch scrapOption {
	case "HOTEL":
		match, _ := regexp.MatchString(tripAdvisorHotelURLRegexp.String(), url)
		return match
	case "RESTAURANT":
		match, _ := regexp.MatchString(tripAdvisorRestaurantRegexp.String(), url)
		return match
	case "AIRLINE":
		match, _ := regexp.MatchString(tripAdvisorAirlineRegexp.String(), url)
		return match
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
func ParseTime(timeToParse string) string {
	// Parse the time string
	t, err := time.Parse(time.RFC3339Nano, timeToParse)
	ErrorHandler(err)

	// Format the time string in a more readable way
	formattedTime := t.Format("01/02/2006 15:04:05 MST")

	return formattedTime
}

// sortStructByTime sorts R2Obj struct by time (newest first)
func sortStructByTime(R2Obj []R2Obj) []R2Obj {

	// Define the comparator function
	less := func(i, j int) bool {

		t1, err := time.Parse(time.RFC3339Nano, R2Obj[i].LastModified)
		if err != nil {
			return false // error handling
		}

		t2, err := time.Parse(time.RFC3339Nano, R2Obj[j].LastModified)
		if err != nil {
			return false // error handling
		}
		return t2.Before(t1)
	}

	// Sort the logs using the comparator function
	sort.Slice(R2Obj, less)

	return R2Obj
}
