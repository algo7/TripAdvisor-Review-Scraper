package tripadvisor

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

// endPointURL is the URL to the TripAdvisor GraphQL endpoint.
const endPointURL = "https://www.tripadvisor.com/data/graphql/ids"

// Query is a function that sends a POST request to the TripAdvisor GraphQL endpoint.
func Query() {

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

	// // Loop through the reviews array and print the review text.
	// for _, review := range response[0].Data.Locations[0].ReviewListPage.Reviews {
	// 	log.Println(review.Text)
	// 	log.Println(review.Title)
	// }



	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Writing header to the CSV file
	header := []string{"Title", "Text", "Rating", "Year", "Month", "Day"}
	if err := writer.Write(header); err != nil {
		log.Fatalln("Error writing header to csv:", err)
	}

	// Iterating over the reviews and writing to the CSV file
	for _, review := range response[0].Data.Locations[0].ReviewListPage.Reviews {

		// Parses the date in yyyy-mm-dd format
		row := []string{
			review.Title,
			review.Text,
			strconv.Itoa(review.Rating),
			review.CreatedDate[0:4],
			review.CreatedDate[5:7],
			review.CreatedDate[8:10],
		}
		if err := writer.Write(row); err != nil {
			log.Fatalln("Error writing row to csv:", err)
		}
	}
	// Print the response body to console (or handle it as needed).
	log.Println(resp.StatusCode)
	// log.Println(string(responseBody))

}
