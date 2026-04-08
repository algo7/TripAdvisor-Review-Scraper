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
	AttractionQueryID string = "ef1a9f94012220d3"

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
	LocationID           uint32  `json:"locationId"`
	Offset               uint32  `json:"offset"`
	Filters              Filters `json:"filters"`
	Limit                uint32  `json:"limit"`
	SortType             any     `json:"sortType"`
	SortBy               string  `json:"sortBy"`
	Language             string  `json:"language"`
	DoMachineTranslation bool    `json:"doMachineTranslation"`
	PhotosPerReviewLimit uint32  `json:"photosPerReviewLimit"`
}

type RouteParams struct {
	GeoID    uint32 `json:"geoId"`
	DetailID uint32 `json:"detailId"`
	Offset   any    `json:"offset"`
}

type RouteRequest struct {
	Fragment string      `json:"fragment"`
	Page     string      `json:"page"`
	Params   RouteParams `json:"params"`
}

type RoutesVariables struct {
	RoutesRequest []RouteRequest `json:"routesRequest"`
}

type RoutesPayload struct {
	Variables  RoutesVariables `json:"variables"`
	Extensions Extensions      `json:"extensions"`
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
		LocationID     any `json:"locationId"`
		Location       any `json:"location"`
		FallbackString any `json:"fallbackString"`
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
	ID              int      `json:"id"`
	Status          string   `json:"status"`
	CreatedDate     string   `json:"createdDate"`
	PublishedDate   string   `json:"publishedDate"`
	Rating          int      `json:"rating"`
	PublishPlatform string   `json:"publishPlatform"`
	Title           string   `json:"title"`
	Language        string   `json:"language"`
	Text            string   `json:"text"`
	Username        string   `json:"username"`
	LocationID      int      `json:"locationId"`
	HelpfulVotes    int      `json:"helpfulVotes"`
	Labels          []string `json:"labels"`
	PhotoIds        []int    `json:"photoIds"`
	TripInfo        struct {
		StayDate string `json:"stayDate"`
		TripType string `json:"tripType"`
	} `json:"tripInfo"`
	Location    ReviewLocation    `json:"location"`
	UserProfile ReviewUserProfile `json:"userProfile"`
}

type ReviewLocation struct {
	LocationID            int    `json:"locationId"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	PlaceType             string `json:"placeType"`
	AccommodationCategory string `json:"accommodationCategory"`
}

type ReviewUserProfile struct {
	ID          string `json:"id"`
	IsVerified  bool   `json:"isVerified"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	Hometown    struct {
		LocationID     any `json:"locationId"`
		Location       any `json:"location"`
		FallbackString any `json:"fallbackString"`
	} `json:"hometown"`
	Route struct {
		URL string `json:"url"`
	} `json:"route"`
	Avatar struct {
		Typename string `json:"__typename"`
		Data     struct {
			ID               int `json:"id"`
			PhotoSizeDynamic struct {
				URLTemplate string `json:"urlTemplate"`
				MaxHeight   int    `json:"maxHeight"`
				MaxWidth    int    `json:"maxWidth"`
			} `json:"photoSizeDynamic"`
		} `json:"data"`
	} `json:"avatar"`
	ContributionCounts struct {
		SumAllUgc int `json:"sumAllUgc"`
	} `json:"contributionCounts"`
}

// ReviewSummary is a struct that represents the review summary object in the response body from TripAdvisor endpoints
type ReviewSummary struct {
	Rating float32 `json:"rating"`
	Count  int     `json:"count"`
}

// ReviewAggregations is a struct that represents the review aggregations object in the response body from TripAdvisor endpoints
type ReviewAggregations struct {
	RatingCounts     []int          `json:"ratingCounts"`
	LanguageCounts   map[string]int `json:"languageCounts"`
	AlertStatusCount int            `json:"alertStatusCount"`
}

// Location is a struct that represents the location object in the response body from TripAdvisor endpoints
type Location struct {
	LocationID            int    `json:"locationId"`
	ParentGeoID           int    `json:"parentGeoId"`
	PlaceType             string `json:"placeType"`
	ReviewSummary         ReviewSummary
	AccommodationCategory string             `json:"accommodationCategory"`
	ReviewAggregations    ReviewAggregations `json:"reviewAggregations"`
}

// Feedback is a struct that represents the feedback object in the response body from TripAdvisor endpoints
type Feedback struct {
	Reviews []Review `json:"reviews"`
}

// Response is a struct that represents the response body from TripAdvisor endpoints
type Response struct {
	Data struct {
		ReviewsProxy []struct {
			TotalCount        int      `json:"totalCount"`
			Reviews           []Review `json:"reviews"`
			ReviewListOptions struct {
				SortType string `json:"sortType"`
				SortBy   string `json:"sortBy"`
			} `json:"reviewListOptions"`
		} `json:"ReviewsProxy_getReviewListPageForLocation"`
	} `json:"data"`
}

// Responses is a slice of Response structs
type Responses []Response
