package tripadvisor

import "regexp"

const (

	// EndPointURL is the URL to the TripAdvisor GraphQL endpoint
	EndPointURL string = "https://www.tripadvisor.com/data/graphql/ids"

	// HotelQueryID is the pre-registered query ID for hotel reviews
	HotelQueryID string = "b83d781ada1db6f2"

	// AirlineQueryID is the pre-registered query ID for airline reviews
	AirlineQueryID string = "83003f8d5a7b1762"

	// AttractionQueryID is the pre-registered query ID for attraction reviews
	AttractionQueryID string = "b83d781ada1db6f2"

	// ReviewLimit is the maximum number of reviews that can be fetched in a single request
	ReviewLimit uint32 = 20
)

var (
	tripAdvisorHotelURLRegexp   = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Hotel_Review-g\d{6,10}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`)
	tripAdvisorRestaurantRegexp = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Restaurant_Review-g\d{6,10}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`)
	tripAdvisorAirlineRegexp    = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Airline_Review-d\d{6,10}-Reviews-[\w-]{1,255}$`)
	tripAdvisorAttractionRegexp = regexp.MustCompile(`^https:\/\/www\.tripadvisor\.com\/Attraction_Review-g\d{6,10}-d\d{1,10}-Reviews-[\w-]{1,255}\.html$`)
)

// Filter is a struct that represents the filter object in the request body to TripAdvisor endpoints
type Filter struct {
	Axis       string   `json:"axis"`
	Selections []string `json:"selections"`
}

// Filters is a slice of Filter structs.
type Filters []Filter

// Variables is a struct that represents the variables object in the request body to TripAdvisor endpoints
type Variables struct {
	LocationID   uint32  `json:"locationId"`
	Offset       uint32  `json:"offset"`
	Prefs        any     `json:"prefs"` // null
	Filters      Filters `json:"filters"`
	InitialPrefs struct {
	} `json:"initialPrefs"` // empty struct
	Limit          uint32 `json:"limit"`
	FilterCacheKey any    `json:"filterCacheKey"` // null
	PrefsCacheKey  string `json:"prefsCacheKey"`  // format "locationReviewPrefs_<localtionID>"
	NeedKeywords   bool   `json:"needKeywords"`   // false
	KeywordVariant string `json:"keywordVariant"`
}

// Extensions is a struct that represents the extensions object in the request body to TripAdvisor endpoints.
type Extensions struct {
	PreRegisteredQueryID string `json:"preRegisteredQueryId"`
}

// Request is a struct that represents the request body to query TripAdvisor endpoints
type Request struct {
	Variables  Variables  `json:"variables"`
	Extensions Extensions `json:"extensions"`
}

// Requests is a slice of Request structs
type Requests []Request

// Photo is a struct that represents the photo object in the response body from TripAdvisor endpoints
type Photo struct {
	ID         int      `json:"id"`
	Statuses   []string `json:"statuses"`
	PhotoSizes []struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		URL    string `json:"url"`
	} `json:"photoSizes"`
}

// Photos is a slice of Photo structs
type Photos []Photo

// UserProfile is a struct that represents the user profile object in the response body from TripAdvisor endpoints
type UserProfile struct {
	IsMe               bool `json:"isMe"`
	IsVerified         bool `json:"isVerified"`
	ContributionCounts struct {
		SumAllUgc   int `json:"sumAllUgc"`
		SumAllLikes int `json:"sumAllLikes"`
	} `json:"contributionCounts"`
	IsFollowing bool   `json:"isFollowing"`
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	Hometown    struct {
		LocationID     interface{} `json:"locationId"`
		Location       interface{} `json:"location"`
		FallbackString interface{} `json:"fallbackString"`
	} `json:"hometown"`
	Route struct {
		URL string `json:"url"`
	} `json:"route"`
	Avatar struct {
		ID         int `json:"id"`
		PhotoSizes []struct {
			Width  int    `json:"width"`
			Height int    `json:"height"`
			URL    string `json:"url"`
		} `json:"photoSizes"`
	} `json:"avatar"`
}

// Review is a struct that represents the review object in the response body from TripAdvisor endpoints
type Review struct {
	ID              int    `json:"id"`
	CreatedDate     string `json:"createdDate"`
	PublishedDate   string `json:"publishedDate"`
	Rating          int    `json:"rating"`
	PublishPlatform string `json:"publishPlatform"`
	TripInfo        struct {
		StayDate string `json:"stayDate"`
		TripType string `json:"tripType"`
	} `json:"tripInfo"`
	PhotoIds    []int       `json:"photoIds"`
	LocationID  int         `json:"locationId"`
	Labels      []string    `json:"labels"`
	Title       string      `json:"title"`
	Text        string      `json:"text"`
	Url         string      `json:"url"`
	Photos      Photos      `json:"photos"`
	UserProfile UserProfile `json:"userProfile"`
	Username    string      `json:"username"`
}

type ReviewSummary struct {
	Rating int `json:"rating"`
	Count  int `json:"count"`
}

type ReviewAggregations struct {
	RatingCounts     []int          `json:"ratingCounts"`
	LanguageCounts   map[string]int `json:"languageCounts"`
	AlertStatusCount int            `json:"alertStatusCount"`
}

type Location struct {
	LocationID            int    `json:"locationId"`
	ParentGeoID           int    `json:"parentGeoId"`
	PlaceType             string `json:"placeType"`
	ReviewSummary         ReviewSummary
	AccommodationCategory string             `json:"accommodationCategory"`
	ReviewAggregations    ReviewAggregations `json:"reviewAggregations"`
}

type Feedback struct {
	Location `json:"location"`
	Reviews  []Review `json:"reviews"`
}

// Response is a struct that represents the response body from TripAdvisor endpoints
type Response struct {
	Data struct {
		Locations []struct {
			Location
			ReviewListPage struct {
				TotalCount int      `json:"totalCount"`
				Reviews    []Review `json:"reviews"`
			} `json:"reviewListPage"`
		} `json:"locations"`
	} `json:"data,omitempty"`
}

// Responses is a slice of Response structs
type Responses []Response
