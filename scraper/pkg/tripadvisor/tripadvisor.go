package tripadvisor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// endPointURL is the URL to the TripAdvisor GraphQL endpoint.
const endPointURL = "https://www.tripadvisor.com/data/graphql/ids"

// Query is a function that sends a POST request to the TripAdvisor GraphQL endpoint.
func Query() error {

	requestFilter := Filter{
		Axis:       "LANGUAGE",
		Selections: []string{"en"},
	}

	requestVariables := Variables{
		LocationID:     8729141,
		Offset:         0,
		Filters:        Filters{requestFilter},
		Limit:          1,
		NeedKeywords:   false,
		PrefsCacheKey:  "locationReviewPrefs_8729141",
		KeywordVariant: "location_keywords_v2_llr_order_30_en",
		InitialPrefs:   struct{}{},
		FilterCacheKey: nil,
		Prefs:          nil,
	}

	requestExtensions := Extensions{
		PreRegisteredQueryID: "b83d781ada1db6f2",
	}

	requestPayload := Request{
		Variables:  requestVariables,
		Extensions: requestExtensions,
	}

	request := Requests{requestPayload}

	// Marshal the request body into JSON.
	// jsonPayload, err := json.Marshal(request)
	// if err != nil {
	// 	log.Fatal("Error marshalling request body: ", err)
	// }

	// Serialize requestPayload to JSON with indentation for pretty printing
	jsonPayload, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		log.Fatal("Error marshalling request payload: ", err)
	}

	// Print pretty JSON payload (optional)
	log.Println(string(jsonPayload))

	// Create a new request using http.NewRequest, setting the method to POST.
	req, err := http.NewRequest("POST", endPointURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatal("Error creating request: ", err)
	}

	// Set the necessary headers as per the original Axios request.
	req.Header.Set("Origin", "https://www.tripadvisor.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.101 Safari/537.36")
	req.Header.Set("X-Requested-By", "someone-special")
	req.Header.Set("Cookie", "asdasdsa")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	// Send the request using an http.Client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request: ", err)
	}
	defer resp.Body.Close()

	// Read the response body.
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body: ", err)
	}

	// Marshal the response body into the Response struct.
	response := Response{}
	err = json.Unmarshal(responseBody, &response)

	// Marshal the response body into JSON to pretty print it with ident.
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal("Error marshalling response body: ", err)
	}

	log.Println(string(jsonResponse))

	// Create a file to save the CSV data
	fileName := "reviews.csv"
	// Create a file to save the CSV data
	fileHandle, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Error creating file %s: %v", fileName, err)
	}

	// Defer closing the file until the function returns
	defer fileHandle.Close()

	// Write the reviews to the CSV file
	headers := []string{"Title", "Text", "Rating", "Year", "Month", "Day"}
	err = WriteReviewToCSV(fileHandle, headers, response[0].Data.Locations[0].ReviewListPage.Reviews)
	if err != nil {
		return fmt.Errorf("Error writing reviews to %s file: %w", fileName, err)
	}

	return nil
}
