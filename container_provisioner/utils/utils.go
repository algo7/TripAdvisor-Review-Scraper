package utils

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Creds struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	AccountId       string `json:"accountId"`
	BucketName      string `json:"bucketName"`
}

// ErrorHandler is a generic error handler
func ErrorHandler(err error) {
	if err != nil {
		formattedError := fmt.Errorf("Error: %w", err)
		fmt.Println(formattedError)
		panic(err)
	}
}

// WriteToFileFromTarStream writes a file to disk
func WriteToFileFromTarStream(fileSuffix string, tarF io.ReadCloser) string {

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
	tarHeader, err := tarReader.Next()
	ErrorHandler(err)

	// Remove the _0 from the file name => the scraper adds the number based on the number of hotels
	// But with the provisioner we do 1 hotel at a time
	newFileName := strings.Replace(tarHeader.Name, "0_", "", 1)
	fileNameToWrite := strings.Replace(newFileName, ".csv", "_"+fileSuffix+".csv", 1)

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

// ParseCreds parses the credentials from a JSON file
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

// GetHotelNameFromURL get the hotel name from the URL
func GetHotelNameFromURL(url string) string {
	// Split the url by "/"
	splitURL := strings.Split(url, "-")

	// Get the last element of the array
	fileName := splitURL[4]

	return fileName
}

// ValidateTripAdvisorURL validates the TripAdvisor Hotel URL
func ValidateTripAdvisorHotelURL(url string) bool {
	regex := `^https:\/\/www\.tripadvisor\.com\/Hotel_Review-g\d{6}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`
	match, _ := regexp.MatchString(regex, url)
	return match
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
	t, err := time.Parse("2006-01-02T15:04:05.000Z", timeToParse)
	ErrorHandler(err)

	// Format the time string in a more readable way
	formattedTime := t.Format("01/02/2006 15:04:05 MST")

	fmt.Println(formattedTime)

	return formattedTime
}

// sortStructByTime sorts R2Obj struct by time
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
		return t1.Before(t2)
	}

	// Sort the logs using the comparator function
	sort.Slice(R2Obj, less)

	return R2Obj
}
