package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
)

// CheckIP takes in a http client and calls ipinfo.io/ip to check the current IP address
func CheckIP(client *http.Client) (ip string, err error) {

	// Make the request to ipinfo.io/ip
	resp, err := client.Get("https://ipinfo.io/ip")
	if err != nil {
		return "", fmt.Errorf("error getting IP address: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response status code: %d", resp.StatusCode)
	}

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(responseBody), nil
}

// GenerateRequestedByID generates a random X-Requested-By ID which is 180 bytes long in ASCII
func GenerateRequestedByID() (requestedByID string, err error) {
	// Define the safe printable ASCII set
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Allocate space for 180 characters
	b := make([]byte, 180)

	// Fill with crypto-random values
	_, err = io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", fmt.Errorf("error generating X-Requested-By ID: %w", err)
	}

	// Map each random byte into the printable set
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}

	return string(b), nil // exactly 180 printable ASCII chars
}
