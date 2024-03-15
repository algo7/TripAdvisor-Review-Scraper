package utils

import (
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
