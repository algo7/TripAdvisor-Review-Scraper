package tripadvisor

// Request is a struct that represents the request body to query TripAdvisor endpoints.
type Request []struct {
	Variables struct {
		LocationID int `json:"locationId"`
		Offset     int `json:"offset"`
		Filters    []struct {
			Axis       string   `json:"axis"`       // "LANGUAGE"
			Selections []string `json:"selections"` // ["en"]
		} `json:"filters"`
		Prefs        any `json:"prefs"` // null
		InitialPrefs struct {
		} `json:"initialPrefs"` // empty struct
		Limit          int    `json:"limit"`
		FilterCacheKey any    `json:"filterCacheKey"` // null
		PrefsCacheKey  string `json:"prefsCacheKey"`  // format "locationReviewPrefs_<localtionID>"
		NeedKeywords   bool   `json:"needKeywords"`   // false
		KeywordVariant string `json:"keywordVariant"`
	} `json:"variables"`
	Extensions struct {
		// Hotel, Restaurant, Airline, etc all have different pre-registered query IDs.
		PreRegisteredQueryID string `json:"preRegisteredQueryId"`
	} `json:"extensions"`
}

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
