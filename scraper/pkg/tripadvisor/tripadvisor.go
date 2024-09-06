package tripadvisor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MakeRequest is a function that sends a POST request to the TripAdvisor GraphQL endpoint
func MakeRequest(client *http.Client, queryID string, language []string, locationID uint32, offset uint32, limit uint32) (responses *Responses, err error) {

	/*
	* Prepare the request body
	 */
	requestFilter := Filter{
		Axis:       "LANGUAGE",
		Selections: language,
	}

	requestVariables := Variables{
		LocationID:     locationID,
		Offset:         offset,
		Filters:        Filters{requestFilter},
		Limit:          limit,
		NeedKeywords:   false,
		PrefsCacheKey:  fmt.Sprintf("locationReviewPrefs_%d", locationID),
		KeywordVariant: "location_keywords_v2_llr_order_30_en",
		InitialPrefs:   struct{}{},
		FilterCacheKey: nil,
		Prefs:          nil,
	}

	requestExtensions := Extensions{
		PreRegisteredQueryID: queryID,
	}

	requestPayload := Request{
		Variables:  requestVariables,
		Extensions: requestExtensions,
	}

	request := Requests{requestPayload}

	// Marshal the request body into JSON
	jsonPayload, err := json.Marshal(request)
	if err != nil {
		log.Fatal("error marshalling request body: ", err)
	}

	// Create a new request using http.NewRequest, setting the method to POST
	req, err := http.NewRequest(http.MethodPost, EndPointURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set the necessary headers as per the original Axios request
	req.Header.Set("Origin", "https://www.tripadvisor.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.101 Safari/537.36")
	req.Header.Set("X-Requested-By", "someone-special")
	req.Header.Set("Cookie", "asdasdsa")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Check for rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("rate Limit Detected: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("error response status code: %d", resp.StatusCode)
	}

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Marshal the response body into the Response struct
	responseData := Responses{}
	err = json.Unmarshal(responseBody, &responseData)

	// Check for errors
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("Raw respsone:\n%s\n", string(responseBody))
	}

	return &responseData, err
}

// GetQueryID is a function that returns the query ID for the given query type
func GetQueryID(queryType string) (queryID string) {

	switch queryType {
	case "HOTEL":
		return HotelQueryID
	case "AIRLINE":
		return AirlineQueryID
	case "ATTRACTION":
		return AttractionQueryID
	default:
		return HotelQueryID
	}
}

// FetchReviewCount is a function that fetches the review count for the given location ID and query type
func FetchReviewCount(client *http.Client, locationID uint32, queryType string, languages []string) (reviewCount int, err error) {

	// Get the query ID for the given query type.
	queryID := GetQueryID(queryType)

	// Make the request to the TripAdvisor GraphQL endpoint.
	responses, err := MakeRequest(client, queryID, languages, locationID, 0, 1)
	if err != nil {
		return 0, fmt.Errorf("error making request: %w", err)
	}

	// Check if responses is nil before dereferencing
	if responses == nil {
		return 0, fmt.Errorf("received nil response for location ID %d", locationID)
	}

	// Now it's safe to dereference responses
	response := *responses

	if len(response) > 0 && len(response[0].Data.Locations) > 0 {
		reviewCount = response[0].Data.Locations[0].ReviewListPage.TotalCount
		return reviewCount, nil
	}

	return 0, fmt.Errorf("no reviews found for location ID %d", locationID)
}

// CalculateIterations is a function that calculates the number of iterations required to fetch all reviews
func CalculateIterations(reviewCount uint32) (iterations uint32) {

	// Calculate the number of iterations required to fetch all reviews
	iterations = reviewCount / ReviewLimit

	// If the review count is not a multiple of ReviewLimit, add one more iteration
	if reviewCount%ReviewLimit != 0 {
		return iterations + 1
	}

	return iterations
}

// CalculateOffset is a function that calculates the offset for the given iteration
func CalculateOffset(iteration uint32) (offset uint32) {
	// Calculate the offset for the given iteration
	offset = iteration * ReviewLimit
	return offset
}

// GetURLType is a function that validates the URL and returns the type of URL
func GetURLType(url string) string {
	if tripAdvisorHotelURLRegexp.MatchString(url) {
		return "HOTEL"
	}

	if tripAdvisorRestaurantRegexp.MatchString(url) {
		return "RESTO"
	}

	if tripAdvisorAirlineRegexp.MatchString(url) {
		return "AIRLINE"
	}

	if tripAdvisorAttractionRegexp.MatchString(url) {
		return "ATTRACTION"
	}

	return ""
}

// ParseURL is a function that parses the URL and returns the location ID and the location name
func ParseURL(url string, locationType string) (locationID uint32, locationName string, error error) {
	// Sample hotel url: https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne_Canton_of_Vaud.html
	// Sample restaurant url: https://www.tripadvisor.com/Restaurant_Review-g187265-d11827759-Reviews-La_Terrasse-Lyon_Rhone_Auvergne_Rhone_Alpes.html
	// Sample airline url: https://www.tripadvisor.com/Airline_Review-d8728979-Reviews-Pegasus-Airlines
	// Sample attraction url: https://www.tripadvisor.com/Attraction_Review-g187261-d195616-Reviews-Mont_Blanc-Chamonix_Haute_Savoie_Auvergne_Rhone_Alpes.html

	switch locationType {

	case "HOTEL", "RESTO", "ATTRACTION":

		// Split the URL by -
		urlSplit := strings.Split(url, "-")

		// Trim the d from the location ID
		locationID, err := strconv.ParseUint(strings.TrimLeft(urlSplit[2], "d"), 10, 32)
		if err != nil {
			return 0, "", fmt.Errorf("error parsing location ID: %w", err)
		}

		// Extract the location name from the URL
		locationName = urlSplit[4]

		return uint32(locationID), locationName, nil

	case "AIRLINE":

		urlSplit := strings.Split(url, "-")
		locationID, err := strconv.ParseUint(strings.TrimLeft(urlSplit[1], "d"), 10, 32)
		if err != nil {
			return 0, "", fmt.Errorf("error parsing location ID: %w", err)
		}

		locationName = strings.Join(urlSplit[3:], "_")

		return uint32(locationID), locationName, nil
	default:
		return 0, "", fmt.Errorf("invalid location type: %s", locationType)
	}
}

// Puts the current reviews into a JSON file
func WriteReviewsToJSONFile(reviews []Review, location Location, fileHandle *os.File) error {
	feedback := Feedback{
		Location: location,
		Reviews:  reviews,
	}
	data, err := json.Marshal(feedback)
	if err != nil {
		return fmt.Errorf("could not marshal data: %w", err)
	}
	if _, err := fileHandle.Write(data); err != nil {
		return fmt.Errorf("could not write data to file: %w", err)
	}
	return nil
}

// SortReviewsByDate is a function that sorts the reviews by date
// This function modifies the original slice
func SortReviewsByDate(reviews []Review) {
	const layout = "2006-01-02" // Move the layout constant here to keep it scoped to the sorting logic
	sort.Slice(reviews, func(i, j int) bool {
		iTime, _ := time.Parse(layout, reviews[i].CreatedDate) // Assume error handling is done elsewhere or errors are unlikely
		jTime, _ := time.Parse(layout, reviews[j].CreatedDate)
		return iTime.After(jTime)
	})

}
