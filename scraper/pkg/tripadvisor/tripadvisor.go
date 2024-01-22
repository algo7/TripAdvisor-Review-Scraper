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
