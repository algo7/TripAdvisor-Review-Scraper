package utils

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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
		os.Exit(0)
	}
}

// WriteToFileFromTarStream writes a file to disk
func WriteToFileFromTarStream(tarF io.ReadCloser) string {

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

	// Create the file
	out, err := os.Create(tarHeader.Name)
	ErrorHandler(err)
	defer out.Close()

	// Write the file to disk
	_, err = io.Copy(out, tarReader)
	ErrorHandler(err)

	// Return the file name
	return tarHeader.Name
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
