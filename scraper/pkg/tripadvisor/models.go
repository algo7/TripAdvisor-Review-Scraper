package tripadvisor

const (
	// HotelQueryID is the pre-registered query ID for hotel reviews
	HotelQueryID = "b83d781ada1db6f2"
	// AirlineQueryID is the pre-registered query ID for airline reviews
	AirlineQueryID = "83003f8d5a7b1762"
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

// Review is a slice of Review structs
type Review struct {
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
}

// Reviews is a slice of Review structs
type Reviews []Review

// Response is a struct that represents the response body from TripAdvisor endpoints
type Response []struct {
	Data struct {
		Locations []struct {
			ReviewListPage struct {
				TotalCount int     `json:"totalCount"`
				Reviews    Reviews `json:"reviews"`
			} `json:"reviewListPage"`
		} `json:"locations"`
	} `json:"data,omitempty"`
}
