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

	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/utils"
)

// MakeRequest is a function that sends a POST request to the TripAdvisor GraphQL endpoint
func MakeRequest(client *http.Client, queryID string, queryType string, language []string, locationID uint32, geoId uint32, offset uint32, limit uint32) (responses *Responses, err error) {

	/*
	* Prepare the request body
	 */
	requestFilter := Filter{
		Axis:       "LANGUAGE",
		Selections: language,
	}

	requestExtensions := Extensions{
		PreRegisteredQueryID: queryID,
	}

	var request any

	if queryType == "AIRLINE" {
		request = BatchRequests{{
			Variables: AirlineVariables{
				LocationID:     locationID,
				Offset:         offset,
				Filters:        Filters{requestFilter},
				Limit:          limit,
				NeedKeywords:   true,
				PrefsCacheKey:  fmt.Sprintf("locationReviewPrefs_%d", locationID),
				KeywordVariant: "location_keywords_v2_llr_order_30_en",
				InitialPrefs:   struct{}{},
				FilterCacheKey: nil,
				Prefs:          nil,
			},
			Extensions: requestExtensions,
		}}

	} else {
		requestVariables := Variables{
			LocationID:           locationID,
			Offset:               offset,
			Filters:              Filters{requestFilter},
			Limit:                limit,
			SortType:             nil,
			SortBy:               "SERVER_DETERMINED",
			Language:             language[0],
			DoMachineTranslation: true,
			PhotosPerReviewLimit: 7,
		}

		routeOffsets := []any{0} // first: number 0
		for i := uint32(1); i <= 7; i++ {
			routeOffsets = append(routeOffsets, fmt.Sprintf("r%d", i*ReviewLimit)) // rest: "r10", "r20"...
		}

		var pageName string
		switch queryType {
		case "HOTEL":
			pageName = "Hotel_Review"
		case "ATTRACTION":
			pageName = "Attraction_Review"
		case "RESTO":
			pageName = "Restaurant_Review"
		case "AIRLINE":
			pageName = "Airline_Review"
		}

		var routes []RouteRequest
		for _, off := range routeOffsets {
			routes = append(routes, RouteRequest{
				Fragment: "",
				Page:     pageName, // adjust based on queryType
				Params: RouteParams{
					GeoID:    geoId, // you'll need geoId passed in or parsed from URL
					DetailID: locationID,
					Offset:   off,
				},
			})
		}

		if queryType == "RESTO" {
			// Batch both into a single request array
			request = BatchRequests{
				{
					Variables:  requestVariables,
					Extensions: requestExtensions,
				},
				{
					Variables:  RoutesVariables{RoutesRequest: routes},
					Extensions: Extensions{PreRegisteredQueryID: queryID},
				},
				// Query for fetching Michelin reviews for restaurants
				{
					Variables:  MichelinVariables{IDs: []uint32{locationID}},
					Extensions: Extensions{PreRegisteredQueryID: MichelinQueryID},
				},
			}
		} else {
			// Batch both into a single request array
			request = BatchRequests{
				{
					Variables:  requestVariables,
					Extensions: requestExtensions,
				},
				{
					Variables:  RoutesVariables{RoutesRequest: routes},
					Extensions: Extensions{PreRegisteredQueryID: queryID},
				},
			}
		}
	}
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

	requestedById, err := utils.GenerateRequestedByID()
	if err != nil {
		return nil, fmt.Errorf("error generating X-Requested-By ID: %w", err)
	}

	// Set the necessary headers as per the original Axios request
	req.Header.Set("Host", "www.tripadvisor.com")
	req.Header.Set("Origin", "https://www.tripadvisor.com")
	req.Header.Set("Referer", "https://www.tripadvisor.com/Hotels")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.101 Safari/537.36")
	req.Header.Set("X-Requested-By", requestedById)
	req.Header.Set("Cookie", fmt.Sprintf("TAUnique=%s", requestedById))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Accepted-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

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

// ExtractReviews extracts the review slice from API responses,
// handling both the ReviewsProxy path (airlines) and the Locations path (hotels/restaurants/attractions).
func ExtractReviews(responses *Responses) []Review {
	if responses == nil {
		return nil
	}
	for _, resp := range *responses {
		if len(resp.Data.ReviewsProxy) > 0 {
			return resp.Data.ReviewsProxy[0].Reviews
		}
		if len(resp.Data.Locations) > 0 {
			return resp.Data.Locations[0].ReviewListPage.Reviews
		}
	}
	return nil
}

// ExtractTotalCount extracts the total review count from API responses.
func ExtractTotalCount(responses *Responses) int {
	if responses == nil {
		return 0
	}
	for _, resp := range *responses {
		if len(resp.Data.ReviewsProxy) > 0 {
			return resp.Data.ReviewsProxy[0].TotalCount
		}
		if len(resp.Data.Locations) > 0 {
			return resp.Data.Locations[0].ReviewListPage.TotalCount
		}
	}
	return 0
}

// ExtractMichelinInfo extracts Michelin star information from API responses.
// Returns nil if no Michelin data is present.
func ExtractMichelinInfo(responses *Responses) *MichelinInfo {
	if responses == nil {
		return nil
	}
	for _, resp := range *responses {
		if len(resp.Data.Michelin) > 0 {
			m := resp.Data.Michelin[0]
			if m.AwardHeader != "" || len(m.Awards) > 0 {
				return &MichelinInfo{
					AwardHeader:   m.AwardHeader,
					AwardReadMore: m.AwardReadMore,
					Awards:        m.Awards,
					Summaries:     m.Summaries,
				}
			}
		}
	}
	return nil
}

// CSVHeaders returns the CSV column headers.
// When includeMichelin is true, Michelin award columns are appended.
func CSVHeaders(includeMichelin bool) []string {
	headers := []string{"Location Name", "Title", "Text", "Rating", "Year", "Month", "Day"}
	if includeMichelin {
		headers = append(headers, "Michelin Award", "Michelin Year")
	}
	return headers
}

// ReviewToCSVRow converts a single Review into a CSV row.
// When michelin is non-nil, Michelin award columns are appended.
func ReviewToCSVRow(r Review, locationName string, michelin *MichelinInfo) []string {
	row := []string{
		locationName,
		r.Title,
		r.Text,
		strconv.Itoa(r.Rating),
		r.CreatedDate[0:4],
		r.CreatedDate[5:7],
		r.CreatedDate[8:10],
	}
	if michelin != nil {
		names := make([]string, 0, len(michelin.Awards))
		years := make([]string, 0, len(michelin.Awards))
		for _, a := range michelin.Awards {
			names = append(names, a.AwardName)
			years = append(years, a.YearOfAward)
		}
		row = append(row, strings.Join(names, "; "), strings.Join(years, "; "))
	}
	return row
}

// FetchReviewCount fetches the review count for the given location ID and query type.
func FetchReviewCount(client *http.Client, locationID uint32, geoID uint32, queryType string, languages []string) (int, error) {

	// Get the query ID for the given query type.
	queryID := GetQueryID(queryType)

	// Make the request to the TripAdvisor GraphQL endpoint.
	responses, err := MakeRequest(client, queryID, queryType, languages, locationID, geoID, 0, 1)
	if err != nil {
		return 0, fmt.Errorf("error making request: %w", err)
	}

	count := ExtractTotalCount(responses)
	if count == 0 {
		return 0, fmt.Errorf("no reviews found for location ID %d", locationID)
	}

	return count, nil
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
func ParseURL(url string, locationType string) (locationID uint32, geoID uint32, locationName string, error error) {
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
			return 0, 0, "", fmt.Errorf("error parsing location ID: %w", err)
		}

		geoID, err := strconv.ParseUint(strings.TrimLeft(urlSplit[1], "g"), 10, 32)
		if err != nil {
			return 0, 0, "", fmt.Errorf("error parsing geo ID: %w", err)
		}

		// Extract the location name from the URL
		locationName = urlSplit[4]

		return uint32(locationID), uint32(geoID), locationName, nil

	case "AIRLINE":

		urlSplit := strings.Split(url, "-")
		locationID, err := strconv.ParseUint(strings.TrimLeft(urlSplit[1], "d"), 10, 32)
		if err != nil {
			return 0, 0, "", fmt.Errorf("error parsing location ID: %w", err)
		}

		locationName = strings.Join(urlSplit[3:], "_")

		return uint32(locationID), 0, locationName, nil
	default:
		return 0, 0, "", fmt.Errorf("invalid location type: %s", locationType)
	}
}

// WriteScrapeResultToJSONFile writes a ScrapeResult (reviews + optional Michelin data) to a JSON file.
func WriteScrapeResultToJSONFile(result *ScrapeResult, fileHandle *os.File) error {
	encoder := json.NewEncoder(fileHandle)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
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
