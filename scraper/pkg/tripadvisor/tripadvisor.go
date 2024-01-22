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

type Filter struct {
	Axis       string   `json:"axis"`
	Selections []string `json:"selections"`
}

type Filters []Filter

type Variables struct {
	LocationID   int     `json:"locationId"`
	Offset       int     `json:"offset"`
	Prefs        any     `json:"prefs"` // null
	Filters      Filters `json:"filters"`
	InitialPrefs struct {
	} `json:"initialPrefs"` // empty struct
	Limit          int    `json:"limit"`
	FilterCacheKey any    `json:"filterCacheKey"` // null
	PrefsCacheKey  string `json:"prefsCacheKey"`  // format "locationReviewPrefs_<localtionID>"
	NeedKeywords   bool   `json:"needKeywords"`   // false
	KeywordVariant string `json:"keywordVariant"`
}

type Extensions struct {
	PreRegisteredQueryID string `json:"preRegisteredQueryId"`
}

// Request is a struct that represents the request body to query TripAdvisor endpoints.
type Request struct {
	Variables  Variables  `json:"variables"`
	Extensions Extensions `json:"extensions"`
}

type Requests []Request

// Response is a struct that represents the response body from TripAdvisor endpoints.
type Response []struct {
	Data struct {
		Locations []struct {
			ReviewListPage struct {
				TotalCount int `json:"totalCount"`
				Reviews    []struct {
					CreatedDate     string `json:"createdDate"`
					PublishedDate   string `json:"publishedDate"`
					Rating          int    `json:"rating"`
					PublishPlatform string `json:"publishPlatform"`
					TripInfo        struct {
						StayDate string `json:"stayDate"`
						TripType string `json:"tripType"`
					} `json:"tripInfo"`
					LocationID int      `json:"locationId"`
					Labels     []string `json:"labels"`
					Title      string   `json:"title"`
					Text       string   `json:"text"`
				} `json:"reviews"`
			} `json:"reviewListPage"`
		} `json:"locations"`
	} `json:"data,omitempty"`
}

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
	req.Header.Set("Referer", "https://www.tripadvisor.com/Airline_Review-d8729141-Reviews-or20-Ryanair.html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.101 Safari/537.36")
	req.Header.Set("X-Requested-By", "bpo-request")
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

	// Create a file to save the CSV data
	file, err := os.Create("reviews.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

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
