package tripadvisor

import (
	"bytes"
	"encoding/json"
	"errors"
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

// MaxRequestAttempts is how many times a request is sent before giving up on a transient failure.
const MaxRequestAttempts = 3

// retryBackoff is the wait before the first retry. It doubles for every further retry.
var retryBackoff = 2 * time.Second

// transientError marks a failure that may succeed on a new attempt, such as a dropped connection or a 5xx response.
type transientError struct{ err error }

func (e transientError) Error() string { return e.err.Error() }
func (e transientError) Unwrap() error { return e.err }

// isTransient reports whether err is worth retrying.
func isTransient(err error) bool {
	var transient transientError
	return errors.As(err, &transient)
}

// MakeRequest is a function that sends a POST request to the TripAdvisor GraphQL endpoint.
// language is a single language tag (e.g. "en"); an empty string requests reviews in all languages.
// Transient failures (dropped connections, timeouts, 5xx responses) are retried up to MaxRequestAttempts times.
// Rate limiting and GraphQL errors are returned straight away.
func MakeRequest(client *http.Client, queryID string, queryType string, language string, locationID uint32, offset uint32, limit uint32) (responses *Responses, err error) {

	request := BuildRequest(queryID, queryType, language, locationID, offset, limit)

	// Marshal the request body into JSON
	jsonPayload, err := json.Marshal(request)
	if err != nil {
		log.Fatal("error marshalling request body: ", err)
	}

	delay := retryBackoff
	for attempt := 1; ; attempt++ {
		responses, err = sendRequest(client, jsonPayload)
		if err == nil || !isTransient(err) {
			return responses, err
		}
		if attempt == MaxRequestAttempts {
			return nil, fmt.Errorf("giving up after %d attempts: %w", attempt, err)
		}
		log.Printf("Request attempt %d/%d failed: %v. Retrying in %v", attempt, MaxRequestAttempts, err, delay)
		time.Sleep(delay)
		delay *= 2
	}
}

// sendRequest sends the GraphQL request once. Failures worth retrying are wrapped in transientError.
func sendRequest(client *http.Client, jsonPayload []byte) (*Responses, error) {

	// Create a new request using http.NewRequest, setting the method to POST.
	// A new reader is used on every call so a retry sends the full body again.
	req, err := http.NewRequest(http.MethodPost, EndPointURL, bytes.NewReader(jsonPayload))
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
		return nil, transientError{fmt.Errorf("error sending request: %w", err)}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Check for rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("rate Limit Detected: %d", resp.StatusCode)
		}
		// Server errors are usually temporary, other client errors are not
		if resp.StatusCode >= http.StatusInternalServerError {
			return nil, transientError{fmt.Errorf("error response status code: %d", resp.StatusCode)}
		}
		return nil, fmt.Errorf("error response status code: %d", resp.StatusCode)
	}

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, transientError{fmt.Errorf("error reading response body: %w", err)}
	}

	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("Raw response:\n%s\n", string(responseBody))
	}

	// Marshal the response body into the Response struct
	responseData := Responses{}

	err = json.Unmarshal(responseBody, &responseData)

	// Check for errors
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if err := ResponseError(&responseData); err != nil {
		return nil, err
	}

	return &responseData, nil
}

// BuildRequest builds the batched GraphQL request body for the given location type.
func BuildRequest(queryID string, queryType string, language string, locationID uint32, offset uint32, limit uint32) BatchRequests {

	if queryType == "AIRLINE" {
		selections := []string{}
		if language != "" {
			selections = append(selections, language)
		}
		return BatchRequests{{
			Variables: AirlineVariables{
				LocationID:     locationID,
				Offset:         offset,
				Filters:        Filters{{Axis: "LANGUAGE", Selections: selections}},
				Limit:          limit,
				NeedKeywords:   true,
				PrefsCacheKey:  fmt.Sprintf("locationReviewPrefs_%d", locationID),
				KeywordVariant: "location_keywords_v2_llr_order_30_en",
				InitialPrefs:   struct{}{},
				FilterCacheKey: nil,
				Prefs:          nil,
			},
			Extensions: Extensions{PreRegisteredQueryID: queryID},
		}}
	}

	filterOptions := FilterOptions{}
	if language != "" {
		filterOptions.Language = &LanguageFilter{LanguageTag: language}
	}

	request := BatchRequests{{
		Variables: Variables{
			ClientID:      GetClientID(queryType),
			LocationID:    locationID,
			SortBy:        "SERVER_DETERMINED",
			Limit:         limit,
			Offset:        offset,
			FilterOptions: filterOptions,
			Locale:        Locale,
		},
		Extensions: Extensions{PreRegisteredQueryID: queryID},
	}}

	// Query for fetching Michelin awards for restaurants
	if queryType == "RESTO" {
		request = append(request, BatchRequest{
			Variables:  MichelinVariables{IDs: []uint32{locationID}},
			Extensions: Extensions{PreRegisteredQueryID: MichelinQueryID},
		})
	}

	return request
}

// ResponseError returns an error if any response in the batch carries a GraphQL error
// (e.g. PersistedQueryNotFound when a query ID is retired) or a ReviewsProxy error
// (e.g. INVALID_REQUEST). Without this check such failures look like "no reviews found".
func ResponseError(responses *Responses) error {
	if responses == nil {
		return nil
	}
	for _, resp := range *responses {
		if len(resp.Errors) > 0 {
			msgs := make([]string, 0, len(resp.Errors))
			for _, e := range resp.Errors {
				msgs = append(msgs, e.Message)
			}
			return fmt.Errorf("graphql error: %s", strings.Join(msgs, "; "))
		}
		for _, rp := range resp.Data.ReviewsProxy {
			if rp.Error != nil {
				return fmt.Errorf("reviews API error: %s: %s", rp.Error.Reason, rp.Error.Details)
			}
		}
	}
	return nil
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
	case "RESTO":
		return RestaurantQueryID
	default:
		return HotelQueryID
	}
}

// GetClientID returns the clientId the review query expects for the given query type.
func GetClientID(queryType string) string {
	switch queryType {
	case "ATTRACTION":
		return AttractionClientID
	case "RESTO":
		return RestaurantClientID
	default:
		return HotelClientID
	}
}

// ExtractReviews extracts the review slice from API responses,
// handling both the ReviewsProxy path (hotels/restaurants/attractions) and the Locations path (airlines).
// Reviews from the ReviewsProxy path are normalized so CreatedDate, TripInfo.TripType and
// TripInfo.StayDate are populated the same way as for airlines.
func ExtractReviews(responses *Responses) []Review {
	if responses == nil {
		return nil
	}
	for _, resp := range *responses {
		if len(resp.Data.ReviewsProxy) > 0 {
			reviews := resp.Data.ReviewsProxy[0].Result.Reviews
			for i := range reviews {
				NormalizeReview(&reviews[i])
			}
			return reviews
		}
		if len(resp.Data.Locations) > 0 {
			return resp.Data.Locations[0].ReviewListPage.Reviews
		}
	}
	return nil
}

// NormalizeReview fills the legacy fields (CreatedDate, TripInfo.TripType, TripInfo.StayDate)
// from the fields returned by the ReviewsProxy_getPublishedReviewsForLocation query.
// Fields that are already set are left untouched.
func NormalizeReview(r *Review) {
	if r.CreatedDate == "" && len(r.SubmittedDateTime) >= 10 {
		r.CreatedDate = r.SubmittedDateTime[:10]
	}
	if r.TripInfo.TripType == "" {
		r.TripInfo.TripType = r.TripInfo.PartyType
		// The old API returned "NONE" when the reviewer gave trip details but no trip type;
		// the new API returns a null partyType alongside the stay/visit date.
		if r.TripInfo.TripType == "" && r.TripInfo.StayOrVisitDate != nil {
			r.TripInfo.TripType = "NONE"
		}
	}
	if r.TripInfo.StayDate == "" && r.TripInfo.StayOrVisitDate != nil {
		d := r.TripInfo.StayOrVisitDate
		if d.Year > 0 && d.Month > 0 {
			if d.Day != nil && *d.Day > 0 {
				r.TripInfo.StayDate = fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, *d.Day)
			} else {
				r.TripInfo.StayDate = fmt.Sprintf("%04d-%02d", d.Year, d.Month)
			}
		}
	}
	if r.UserProfile.ID == "" {
		r.UserProfile.ID = r.UserProfile.UserID
	}
	if r.Location.Name == "" && r.LocationNode != nil {
		info := r.LocationNode.Detail.Info
		r.Location.LocationID = r.LocationID
		r.Location.Name = info.LocalizedName
		r.Location.PlaceType = info.PlaceType
		r.Location.URL = info.PrimaryRoute.URL
	}
}

// DedupeReviews returns the reviews whose ID has not been seen yet and records their IDs in seen.
func DedupeReviews(reviews []Review, seen map[int]struct{}) []Review {
	unique := make([]Review, 0, len(reviews))
	for _, r := range reviews {
		if _, ok := seen[r.ID]; ok {
			continue
		}
		seen[r.ID] = struct{}{}
		unique = append(unique, r)
	}
	return unique
}

// ExtractTotalCount extracts the total review count from API responses.
func ExtractTotalCount(responses *Responses) int {
	if responses == nil {
		return 0
	}
	for _, resp := range *responses {
		if len(resp.Data.ReviewsProxy) > 0 {
			return resp.Data.ReviewsProxy[0].Result.TotalCount
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
	headers := []string{"Location Name", "Title", "Text", "Rating", "Year", "Month", "Day", "Trip Type", "Stay Date"}
	if includeMichelin {
		headers = append(headers, "Michelin Award", "Michelin Year")
	}
	return headers
}

// ReviewToCSVRow converts a single Review into a CSV row.
// When michelin is non-nil, Michelin award columns are appended.
func ReviewToCSVRow(r Review, locationName string, michelin *MichelinInfo) []string {
	var year, month, day string
	if len(r.CreatedDate) >= 10 {
		year, month, day = r.CreatedDate[0:4], r.CreatedDate[5:7], r.CreatedDate[8:10]
	}
	row := []string{
		locationName,
		r.Title,
		r.Text,
		strconv.Itoa(r.Rating),
		year,
		month,
		day,
		r.TripInfo.TripType,
		r.TripInfo.StayDate,
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

// FetchReviewCount fetches the review count for the given location ID, query type and language.
// A count of 0 is not an error: a location can have no reviews in a given language.
func FetchReviewCount(client *http.Client, locationID uint32, queryType string, language string) (int, error) {

	// Get the query ID for the given query type.
	queryID := GetQueryID(queryType)

	// Make the request to the TripAdvisor GraphQL endpoint.
	responses, err := MakeRequest(client, queryID, queryType, language, locationID, 0, 1)
	if err != nil {
		return 0, fmt.Errorf("error making request: %w", err)
	}

	return ExtractTotalCount(responses), nil
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
