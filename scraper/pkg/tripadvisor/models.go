package tripadvisor

import "regexp"

const (

	// EndPointURL is the URL to the TripAdvisor GraphQL endpoint
	EndPointURL string = "https://www.tripadvisor.com/data/graphql/ids"

	// HotelQueryID is the pre-registered query ID for hotel reviews
	// (ReviewsProxy_getPublishedReviewsForLocation, shared by hotels, attractions and restaurants)
	HotelQueryID string = "3b6e3b9a31d4be70"

	// AirlineQueryID is the pre-registered query ID for airline reviews
	AirlineQueryID string = "e1ca245af416c316"

	// AttractionQueryID is the pre-registered query ID for attraction reviews
	AttractionQueryID string = "3b6e3b9a31d4be70"

	// RestaurantQueryID is the pre-registered query ID for restaurant reviews
	RestaurantQueryID string = "3b6e3b9a31d4be70"

	// HotelClientID is the clientId the TripAdvisor web app sends when fetching hotel reviews
	HotelClientID string = "HOTEL_DETAIL_LIST"

	// AttractionClientID is the clientId the TripAdvisor web app sends when fetching attraction reviews
	AttractionClientID string = "ATTRACTION_DETAIL_LIST"

	// RestaurantClientID is the clientId the TripAdvisor web app sends when fetching restaurant reviews
	RestaurantClientID string = "RESTAURANT_DETAIL_LIST"

	// Locale is the locale sent with review requests
	Locale string = "en-US"

	// MichelinQueryID is the pre-registered query ID for getting Michelin Star status of restaurants
	MichelinQueryID string = "496720f897546a4e"

	// ReviewLimit is the maximum number of reviews that can be fetched in a single request.
	// The API rejects anything above 20 ("limit must be between 1 and 20").
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

// MichelinVariables is a struct that represents the variables object in the request body to get Michelin Star status of restaurants.
type MichelinVariables struct {
	// List of location IDs of the restaurants to get Michelin Star status for.
	IDs []uint32 `json:"ids"`
}

// AirlineVariables is a struct that represents the variables object in the request body to get airline reviews.
// It's different from the Variables struct used for other types of reviews because the airline reviews endpoint has different requirements for the variables object.
type AirlineVariables struct {
	LocationID     uint32   `json:"locationId"`
	Offset         uint32   `json:"offset"`
	Prefs          any      `json:"prefs"`
	Filters        Filters  `json:"filters"`
	InitialPrefs   struct{} `json:"initialPrefs"`
	Limit          uint32   `json:"limit"`
	FilterCacheKey any      `json:"filterCacheKey"`
	PrefsCacheKey  string   `json:"prefsCacheKey"`
	NeedKeywords   bool     `json:"needKeywords"`
	KeywordVariant string   `json:"keywordVariant"`
}

// LanguageFilter selects reviews written in a single language.
type LanguageFilter struct {
	LanguageTag string `json:"languageTag"`
}

// FilterOptions is the filter object of the review request.
// The API accepts a single language per request; omitting Language returns reviews in all languages.
type FilterOptions struct {
	Language *LanguageFilter `json:"language,omitempty"`
}

// Variables is a struct that represents the variables object in the request body
// to the ReviewsProxy_getPublishedReviewsForLocation query (hotels, attractions, restaurants).
type Variables struct {
	ClientID      string        `json:"clientId"`
	LocationID    uint32        `json:"locationId"`
	SortBy        string        `json:"sortBy"`
	Limit         uint32        `json:"limit"`
	Offset        uint32        `json:"offset"`
	FilterOptions FilterOptions `json:"filterOptions"`
	Locale        string        `json:"locale"`
}

// Extensions is a struct that represents the extensions object in the request body to TripAdvisor endpoints.
type Extensions struct {
	PreRegisteredQueryID string `json:"preRegisteredQueryId"`
}

type BatchRequest struct {
	Variables  any        `json:"variables"`
	Extensions Extensions `json:"extensions"`
}

type BatchRequests []BatchRequest

type ReviewLocation struct {
	LocationID            int    `json:"locationId"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	PlaceType             string `json:"placeType"`
	AccommodationCategory string `json:"accommodationCategory"`
}

type ReviewUserProfile struct {
	ID          string `json:"id"`
	UserID      string `json:"userId,omitempty"`
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

// Review is a struct that represents the review object in the response body from TripAdvisor endpoints
type Review struct {
	ID              int               `json:"id"`
	Status          string            `json:"status"`
	CreatedDate     string            `json:"createdDate"`
	PublishedDate   string            `json:"publishedDate"`
	Rating          int               `json:"rating"`
	PublishPlatform string            `json:"publishPlatform"`
	Title           string            `json:"title"`
	Language        string            `json:"language"`
	Text            string            `json:"text"`
	Username        string            `json:"username"`
	LocationID      int               `json:"locationId"`
	HelpfulVotes    int               `json:"helpfulVotes"`
	Labels          []string          `json:"labels"`
	PhotoIds        []int             `json:"photoIds"`
	TripInfo        TripInfo          `json:"tripInfo"`
	Location        ReviewLocation    `json:"location"`
	UserProfile     ReviewUserProfile `json:"userProfile"`

	// SubmittedDateTime is the ISO 8601 timestamp returned by the
	// ReviewsProxy_getPublishedReviewsForLocation query (hotels, attractions, restaurants).
	// It is normalized into CreatedDate by ExtractReviews.
	SubmittedDateTime string `json:"submittedDateTime,omitempty"`
	// TranslationInfo tells whether the text is a machine translation and of which original language.
	TranslationInfo *TranslationInfo `json:"translationInfo,omitempty"`
	// LocationNode holds the name of the reviewed location, which for attractions
	// can be a sub-product (tour) of the location being scraped.
	LocationNode *ReviewLocationNode `json:"locationNode,omitempty"`
}

// TranslationInfo describes the translation of a review returned by the newer review query.
type TranslationInfo struct {
	TranslationType            string `json:"translationType"`
	OriginalLanguage           string `json:"originalLanguage"`
	MachineTranslationProvider string `json:"machineTranslationProvider,omitempty"`
}

// TripInfo holds the trip details of a review.
// The airline query returns StayDate/TripType directly; the newer review query returns
// PartyType/StayOrVisitDate, which ExtractReviews normalizes into StayDate/TripType.
type TripInfo struct {
	StayDate        string           `json:"stayDate"`
	TripType        string           `json:"tripType"`
	PartyType       string           `json:"partyType,omitempty"`
	StayOrVisitDate *StayOrVisitDate `json:"stayOrVisitDate,omitempty"`
}

// StayOrVisitDate is the structured stay/visit date returned by the newer review query.
// Day is often null.
type StayOrVisitDate struct {
	Year  int  `json:"year"`
	Month int  `json:"month"`
	Day   *int `json:"day"`
}

// ReviewLocationNode is the location node attached to reviews by the newer review query.
type ReviewLocationNode struct {
	Detail struct {
		Info struct {
			LocalizedName string `json:"localizedName"`
			PlaceType     string `json:"placeType"`
			PrimaryRoute  struct {
				URL string `json:"url"`
			} `json:"primaryRoute"`
		} `json:"info"`
	} `json:"detail"`
}

// MichelinAward represents a single Michelin award for a restaurant.
type MichelinAward struct {
	AwardName    string `json:"award_name"`
	AwardTitle   string `json:"award_title"`
	YearOfAward  string `json:"yearOfAward"`
	Description  string `json:"description,omitempty"`
	AwardIconURL string `json:"awardIconUrl,omitempty"`
}

// MichelinSummary represents a summary link from the MICHELIN Guide.
type MichelinSummary struct {
	ExternalURL string `json:"externalUrl"`
	Text        string `json:"text"`
}

// MichelinInfo holds Michelin star information for a restaurant.
type MichelinInfo struct {
	AwardHeader   string            `json:"awardHeader,omitempty"`
	AwardReadMore string            `json:"awardReadMore,omitempty"`
	Awards        []MichelinAward   `json:"awards,omitempty"`
	Summaries     []MichelinSummary `json:"summaries,omitempty"`
}

// ScrapeResult holds the complete output of a scrape operation,
// including reviews and optional Michelin data for restaurants.
type ScrapeResult struct {
	Reviews  []Review      `json:"reviews"`
	Michelin *MichelinInfo `json:"michelin,omitempty"`
}

// ReviewsProxyError is the per-request error object returned by ReviewsProxy_getPublishedReviewsForLocation,
// e.g. {"reason":"INVALID_REQUEST","details":"limit must be between 1 and 20"}.
type ReviewsProxyError struct {
	Reason  string `json:"reason"`
	Details string `json:"details"`
}

// ReviewsProxyResult is the result of ReviewsProxy_getPublishedReviewsForLocation.
type ReviewsProxyResult struct {
	Error  *ReviewsProxyError `json:"error"`
	Result struct {
		TotalCount int      `json:"totalCount"`
		SortBy     string   `json:"sortBy"`
		Reviews    []Review `json:"reviews"`
	} `json:"result"`
}

// LocationReviews is a location entry returned by the airline review query.
type LocationReviews struct {
	LocationID     int `json:"locationId"`
	ReviewListPage struct {
		TotalCount int      `json:"totalCount"`
		Reviews    []Review `json:"reviews"`
	} `json:"reviewListPage"`
}

// ResponseData is the data object of a GraphQL response.
type ResponseData struct {
	// ReviewsProxy is returned for hotels, attractions and restaurants.
	ReviewsProxy []ReviewsProxyResult `json:"ReviewsProxy_getPublishedReviewsForLocation"`
	// Locations is returned for airlines.
	Locations []LocationReviews `json:"locations"`
	// Michelin is returned for restaurants.
	Michelin []MichelinInfo `json:"RestaurantAwards_getRestaurantAwards"`
}

// GraphQLError is a top-level GraphQL error, e.g. {"message":"PersistedQueryNotFound"}.
type GraphQLError struct {
	Message string `json:"message"`
}

// Response is a struct that represents the response body from TripAdvisor endpoints
type Response struct {
	Data   ResponseData   `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// Responses is a slice of Response structs
type Responses []Response
