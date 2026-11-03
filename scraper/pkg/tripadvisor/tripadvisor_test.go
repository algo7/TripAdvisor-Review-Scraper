package tripadvisor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetURLType(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "valid hotel URL",
			url:      "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne_Canton_of_Vaud.html",
			expected: "HOTEL",
		},
		{
			name:     "valid restaurant URL",
			url:      "https://www.tripadvisor.com/Restaurant_Review-g187265-d11827759-Reviews-La_Terrasse-Lyon_Rhone_Auvergne_Rhone_Alpes.html",
			expected: "RESTO",
		},
		{
			name:     "valid airline URL",
			url:      "https://www.tripadvisor.com/Airline_Review-d8729113-Reviews-Lufthansa",
			expected: "AIRLINE",
		},
		{
			name:     "valid attraction URL",
			url:      "https://www.tripadvisor.com/Attraction_Review-g187261-d1008501-Reviews-Les_Ailes_du_Mont_Blanc-Chamonix_Haute_Savoie_Auvergne_Rhone_Alpes.html",
			expected: "ATTRACTION",
		},
		{
			name:     "invalid URL returns empty",
			url:      "https://www.tripadvisor.com/SomethingElse",
			expected: "",
		},
		{
			name:     "empty string returns empty",
			url:      "",
			expected: "",
		},
		{
			name:     "non-tripadvisor URL returns empty",
			url:      "https://www.google.com",
			expected: "",
		},
		{
			name:     "tripadvisor .fr domain returns empty",
			url:      "https://www.tripadvisor.fr/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne_Canton_of_Vaud.html",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetURLType(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		locationType    string
		expectedLocID   uint32
		expectedGeoID   uint32
		expectedLocName string
		expectError     bool
	}{
		{
			name:            "parse hotel URL",
			url:             "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne_Canton_of_Vaud.html",
			locationType:    "HOTEL",
			expectedLocID:   231860,
			expectedGeoID:   188107,
			expectedLocName: "Beau_Rivage_Palace",
		},
		{
			name:            "parse restaurant URL",
			url:             "https://www.tripadvisor.com/Restaurant_Review-g187265-d11827759-Reviews-La_Terrasse-Lyon_Rhone_Auvergne_Rhone_Alpes.html",
			locationType:    "RESTO",
			expectedLocID:   11827759,
			expectedGeoID:   187265,
			expectedLocName: "La_Terrasse",
		},
		{
			name:            "parse attraction URL",
			url:             "https://www.tripadvisor.com/Attraction_Review-g187261-d1008501-Reviews-Les_Ailes_du_Mont_Blanc-Chamonix_Haute_Savoie_Auvergne_Rhone_Alpes.html",
			locationType:    "ATTRACTION",
			expectedLocID:   1008501,
			expectedGeoID:   187261,
			expectedLocName: "Les_Ailes_du_Mont_Blanc",
		},
		{
			name:            "parse airline URL",
			url:             "https://www.tripadvisor.com/Airline_Review-d8729113-Reviews-Lufthansa",
			locationType:    "AIRLINE",
			expectedLocID:   8729113,
			expectedGeoID:   0,
			expectedLocName: "Lufthansa",
		},
		{
			name:            "parse airline URL with hyphenated name",
			url:             "https://www.tripadvisor.com/Airline_Review-d8728979-Reviews-Pegasus-Airlines",
			locationType:    "AIRLINE",
			expectedLocID:   8728979,
			expectedGeoID:   0,
			expectedLocName: "Pegasus_Airlines",
		},
		{
			name:         "invalid location type returns error",
			url:          "https://www.tripadvisor.com/SomethingElse",
			locationType: "INVALID",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locID, geoID, locName, err := ParseURL(tt.url, tt.locationType)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLocID, locID)
			assert.Equal(t, tt.expectedGeoID, geoID)
			assert.Equal(t, tt.expectedLocName, locName)
		})
	}
}

func TestGetQueryID(t *testing.T) {
	tests := []struct {
		name      string
		queryType string
		expected  string
	}{
		{
			name:      "hotel returns hotel query ID",
			queryType: "HOTEL",
			expected:  HotelQueryID,
		},
		{
			name:      "airline returns airline query ID",
			queryType: "AIRLINE",
			expected:  AirlineQueryID,
		},
		{
			name:      "attraction returns attraction query ID",
			queryType: "ATTRACTION",
			expected:  AttractionQueryID,
		},
		{
			name:      "restaurant returns restaurant query ID",
			queryType: "RESTO",
			expected:  RestaurantQueryID,
		},
		{
			name:      "unknown type defaults to hotel query ID",
			queryType: "UNKNOWN",
			expected:  HotelQueryID,
		},
		{
			name:      "empty string defaults to hotel query ID",
			queryType: "",
			expected:  HotelQueryID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetQueryID(tt.queryType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateIterations(t *testing.T) {
	tests := []struct {
		name        string
		reviewCount uint32
		expected    uint32
	}{
		{
			name:        "exact multiple of ReviewLimit",
			reviewCount: ReviewLimit * 3,
			expected:    3,
		},
		{
			name:        "one review over a multiple",
			reviewCount: ReviewLimit*3 + 1,
			expected:    4,
		},
		{
			name:        "single review",
			reviewCount: 1,
			expected:    1,
		},
		{
			name:        "exactly ReviewLimit reviews",
			reviewCount: ReviewLimit,
			expected:    1,
		},
		{
			name:        "zero reviews",
			reviewCount: 0,
			expected:    0,
		},
		{
			name:        "ReviewLimit minus one",
			reviewCount: ReviewLimit - 1,
			expected:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateIterations(tt.reviewCount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		name      string
		iteration uint32
		expected  uint32
	}{
		{
			name:      "first iteration",
			iteration: 0,
			expected:  0,
		},
		{
			name:      "second iteration",
			iteration: 1,
			expected:  ReviewLimit,
		},
		{
			name:      "fifth iteration",
			iteration: 5,
			expected:  ReviewLimit * 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOffset(tt.iteration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCSVHeaders(t *testing.T) {
	tests := []struct {
		name            string
		includeMichelin bool
		expected        []string
	}{
		{
			name:            "without Michelin columns",
			includeMichelin: false,
			expected:        []string{"Location Name", "Title", "Text", "Rating", "Year", "Month", "Day", "Trip Type", "Stay Date"},
		},
		{
			name:            "with Michelin columns",
			includeMichelin: true,
			expected:        []string{"Location Name", "Title", "Text", "Rating", "Year", "Month", "Day", "Trip Type", "Stay Date", "Michelin Award", "Michelin Year"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVHeaders(tt.includeMichelin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReviewToCSVRow(t *testing.T) {
	tests := []struct {
		name         string
		review       Review
		locationName string
		michelin     *MichelinInfo
		expected     []string
	}{
		{
			name: "basic review without Michelin",
			review: Review{
				Title:       "Great stay",
				Text:        "Loved it",
				Rating:      5,
				CreatedDate: "2025-06-15",
				TripInfo: TripInfo{
					StayDate: "2025-06-01",
					TripType: "COUPLES",
				},
			},
			locationName: "Test_Hotel",
			michelin:     nil,
			expected:     []string{"Test_Hotel", "Great stay", "Loved it", "5", "2025", "06", "15", "COUPLES", "2025-06-01"},
		},
		{
			name: "review with Michelin info",
			review: Review{
				Title:       "Fine dining",
				Text:        "Excellent food",
				Rating:      4,
				CreatedDate: "2024-12-25",
				TripInfo: TripInfo{
					StayDate: "2024-12-20",
					TripType: "FAMILY",
				},
			},
			locationName: "Fancy_Restaurant",
			michelin: &MichelinInfo{
				Awards: []MichelinAward{
					{AwardName: "1 Star", YearOfAward: "2024"},
				},
			},
			expected: []string{"Fancy_Restaurant", "Fine dining", "Excellent food", "4", "2024", "12", "25", "FAMILY", "2024-12-20", "1 Star", "2024"},
		},
		{
			name: "review with multiple Michelin awards",
			review: Review{
				Title:       "Amazing",
				Text:        "Best meal ever",
				Rating:      5,
				CreatedDate: "2025-01-10",
				TripInfo: TripInfo{
					StayDate: "2025-01-05",
					TripType: "SOLO",
				},
			},
			locationName: "Star_Restaurant",
			michelin: &MichelinInfo{
				Awards: []MichelinAward{
					{AwardName: "2 Stars", YearOfAward: "2024"},
					{AwardName: "1 Star", YearOfAward: "2023"},
				},
			},
			expected: []string{"Star_Restaurant", "Amazing", "Best meal ever", "5", "2025", "01", "10", "SOLO", "2025-01-05", "2 Stars; 1 Star", "2024; 2023"},
		},
		{
			name: "review with empty trip info",
			review: Review{
				Title:       "OK stay",
				Text:        "Nothing special",
				Rating:      3,
				CreatedDate: "2025-03-20",
			},
			locationName: "Average_Hotel",
			michelin:     nil,
			expected:     []string{"Average_Hotel", "OK stay", "Nothing special", "3", "2025", "03", "20", "", ""},
		},
	}

	tests = append(tests, struct {
		name         string
		review       Review
		locationName string
		michelin     *MichelinInfo
		expected     []string
	}{
		name:         "review with missing date does not panic",
		review:       Review{Title: "No date", Text: "x", Rating: 2},
		locationName: "Hotel",
		expected:     []string{"Hotel", "No date", "x", "2", "", "", "", "", ""},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReviewToCSVRow(tt.review, tt.locationName, tt.michelin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetClientID(t *testing.T) {
	assert.Equal(t, HotelClientID, GetClientID("HOTEL"))
	assert.Equal(t, AttractionClientID, GetClientID("ATTRACTION"))
	assert.Equal(t, RestaurantClientID, GetClientID("RESTO"))
	assert.Equal(t, HotelClientID, GetClientID(""))
}

// reviewsProxyResponse builds a Response carrying a ReviewsProxy_getPublishedReviewsForLocation result.
func reviewsProxyResponse(total int, reviews ...Review) Response {
	rp := ReviewsProxyResult{}
	rp.Result.TotalCount = total
	rp.Result.Reviews = reviews
	return Response{Data: ResponseData{ReviewsProxy: []ReviewsProxyResult{rp}}}
}

// locationsResponse builds a Response carrying an airline locations result.
func locationsResponse(total int, reviews ...Review) Response {
	loc := LocationReviews{LocationID: 12345}
	loc.ReviewListPage.TotalCount = total
	loc.ReviewListPage.Reviews = reviews
	return Response{Data: ResponseData{Locations: []LocationReviews{loc}}}
}

func TestExtractReviews(t *testing.T) {
	day := 7
	tests := []struct {
		name      string
		responses *Responses
		expected  []Review
	}{
		{
			name:      "nil responses returns nil",
			responses: nil,
			expected:  nil,
		},
		{
			name:      "empty responses returns nil",
			responses: &Responses{},
			expected:  nil,
		},
		{
			name:      "extracts from Locations path (airline) unchanged",
			responses: &Responses{locationsResponse(1, Review{ID: 1, Title: "Test Review", CreatedDate: "2026-10-26"})},
			expected:  []Review{{ID: 1, Title: "Test Review", CreatedDate: "2026-10-26"}},
		},
		{
			name: "extracts from ReviewsProxy path and normalizes fields",
			responses: &Responses{reviewsProxyResponse(1, Review{
				ID:                100,
				Title:             "Hotel Review",
				SubmittedDateTime: "2026-10-29T16:03:54.298Z",
				TripInfo:          TripInfo{PartyType: "FRIENDS", StayOrVisitDate: &StayOrVisitDate{Year: 2026, Month: 10, Day: &day}},
				UserProfile:       ReviewUserProfile{UserID: "abc"},
			})},
			expected: []Review{{
				ID:                100,
				Title:             "Hotel Review",
				CreatedDate:       "2026-10-29",
				SubmittedDateTime: "2026-10-29T16:03:54.298Z",
				TripInfo:          TripInfo{TripType: "FRIENDS", StayDate: "2026-10-07", PartyType: "FRIENDS", StayOrVisitDate: &StayOrVisitDate{Year: 2026, Month: 10, Day: &day}},
				UserProfile:       ReviewUserProfile{ID: "abc", UserID: "abc"},
			}},
		},
		{
			name: "ReviewsProxy is found in a later batch entry (after Michelin)",
			responses: &Responses{
				{Data: ResponseData{Michelin: []MichelinInfo{{}}}},
				reviewsProxyResponse(1, Review{ID: 5, SubmittedDateTime: "2025-01-02T00:00:00Z"}),
			},
			expected: []Review{{ID: 5, CreatedDate: "2025-01-02", SubmittedDateTime: "2025-01-02T00:00:00Z"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractReviews(tt.responses)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeReview(t *testing.T) {
	t.Run("stay date without day", func(t *testing.T) {
		r := Review{TripInfo: TripInfo{StayOrVisitDate: &StayOrVisitDate{Year: 2026, Month: 3}}}
		NormalizeReview(&r)
		assert.Equal(t, "2026-03", r.TripInfo.StayDate)
	})
	t.Run("null party type with trip details maps to NONE like the old API", func(t *testing.T) {
		r := Review{TripInfo: TripInfo{StayOrVisitDate: &StayOrVisitDate{Year: 2026, Month: 3}}}
		NormalizeReview(&r)
		assert.Equal(t, "NONE", r.TripInfo.TripType)
	})
	t.Run("missing fields stay empty", func(t *testing.T) {
		r := Review{}
		NormalizeReview(&r)
		assert.Equal(t, "", r.CreatedDate)
		assert.Equal(t, "", r.TripInfo.StayDate)
		assert.Equal(t, "", r.TripInfo.TripType)
	})
	t.Run("existing legacy fields are not overwritten", func(t *testing.T) {
		r := Review{CreatedDate: "2020-01-01", SubmittedDateTime: "2026-10-29T16:03:54.298Z", TripInfo: TripInfo{TripType: "SOLO", PartyType: "FAMILY"}}
		NormalizeReview(&r)
		assert.Equal(t, "2020-01-01", r.CreatedDate)
		assert.Equal(t, "SOLO", r.TripInfo.TripType)
	})
}

func TestDedupeReviews(t *testing.T) {
	seen := map[int]struct{}{}
	first := DedupeReviews([]Review{{ID: 1, Language: "en"}, {ID: 2, Language: "en"}}, seen)
	assert.Equal(t, []int{1, 2}, []int{first[0].ID, first[1].ID})

	// The fr page returns review 1 again (machine translated) plus a new review 3
	second := DedupeReviews([]Review{{ID: 1, Language: "fr"}, {ID: 3, Language: "fr"}}, seen)
	assert.Equal(t, []Review{{ID: 3, Language: "fr"}}, second)
}

func TestExtractTotalCount(t *testing.T) {
	tests := []struct {
		name      string
		responses *Responses
		expected  int
	}{
		{
			name:      "nil responses returns 0",
			responses: nil,
			expected:  0,
		},
		{
			name:      "empty responses returns 0",
			responses: &Responses{},
			expected:  0,
		},
		{
			name:      "extracts count from Locations path",
			responses: &Responses{locationsResponse(42)},
			expected:  42,
		},
		{
			name:      "extracts count from ReviewsProxy path",
			responses: &Responses{reviewsProxyResponse(99)},
			expected:  99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTotalCount(tt.responses)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseError(t *testing.T) {
	assert.NoError(t, ResponseError(nil))
	assert.NoError(t, ResponseError(&Responses{reviewsProxyResponse(1)}))

	persisted := &Responses{{Errors: []GraphQLError{{Message: "PersistedQueryNotFound"}}}}
	err := ResponseError(persisted)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PersistedQueryNotFound")

	invalid := &Responses{{Data: ResponseData{ReviewsProxy: []ReviewsProxyResult{{Error: &ReviewsProxyError{Reason: "INVALID_REQUEST", Details: "limit must be between 1 and 20"}}}}}}
	err = ResponseError(invalid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit must be between 1 and 20")
}

func TestResponseUnmarshal(t *testing.T) {
	// Trimmed-down real response of ReviewsProxy_getPublishedReviewsForLocation (query 3b6e3b9a31d4be70)
	body := `[{"data":{"ReviewsProxy_getPublishedReviewsForLocation":[{"error":null,"result":{"totalCount":3570,"sortBy":"MOST_RECENT","reviews":[{"id":1083670442,"locationId":19587144,"title":"Memorable experience","text":"Great trek","rating":5,"language":"en","helpfulVotes":0,"submittedDateTime":"2026-10-29T16:03:54.298Z","tripInfo":{"partyType":"FRIENDS","stayOrVisitDate":{"year":2026,"month":10,"day":null}},"userProfile":{"userId":"ABC","displayName":"Jo","username":"jo1"},"locationNode":{"detail":{"info":{"localizedName":"Gokyo Valley Trek","placeType":"ATTRACTION_PRODUCT","primaryRoute":{"url":"/x"}}}}}]}}]}}]`
	var responses Responses
	assert.NoError(t, json.Unmarshal([]byte(body), &responses))
	assert.NoError(t, ResponseError(&responses))
	assert.Equal(t, 3570, ExtractTotalCount(&responses))

	reviews := ExtractReviews(&responses)
	assert.Len(t, reviews, 1)
	r := reviews[0]
	assert.Equal(t, "2026-10-29", r.CreatedDate)
	assert.Equal(t, "FRIENDS", r.TripInfo.TripType)
	assert.Equal(t, "2026-10", r.TripInfo.StayDate)
	assert.Equal(t, "Gokyo Valley Trek", r.LocationNode.Detail.Info.LocalizedName)
	assert.Equal(t, ReviewLocation{LocationID: 19587144, Name: "Gokyo Valley Trek", PlaceType: "ATTRACTION_PRODUCT", URL: "/x"}, r.Location)
	assert.Equal(t, []string{"Jo", "Memorable experience", "Great trek", "5", "2026", "10", "29", "FRIENDS", "2026-10"}, ReviewToCSVRow(r, "Jo", nil))
}

func TestBuildRequest(t *testing.T) {
	t.Run("hotel request matches the web app contract", func(t *testing.T) {
		req := BuildRequest(HotelQueryID, "HOTEL", "en", 231860, 20, 20)
		b, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.JSONEq(t, `[{"variables":{"clientId":"HOTEL_DETAIL_LIST","locationId":231860,"sortBy":"SERVER_DETERMINED","limit":20,"offset":20,"filterOptions":{"language":{"languageTag":"en"}},"locale":"en-US"},"extensions":{"preRegisteredQueryId":"3b6e3b9a31d4be70"}}]`, string(b))
	})
	t.Run("empty language omits the language filter", func(t *testing.T) {
		b, _ := json.Marshal(BuildRequest(AttractionQueryID, "ATTRACTION", "", 1, 0, 1))
		assert.Contains(t, string(b), `"filterOptions":{}`)
		assert.Contains(t, string(b), `"clientId":"ATTRACTION_DETAIL_LIST"`)
	})
	t.Run("restaurant request batches the Michelin query", func(t *testing.T) {
		req := BuildRequest(RestaurantQueryID, "RESTO", "en", 11827759, 0, 1)
		assert.Len(t, req, 2)
		assert.Equal(t, MichelinQueryID, req[1].Extensions.PreRegisteredQueryID)
		assert.Equal(t, RestaurantClientID, req[0].Variables.(Variables).ClientID)
	})
	t.Run("airline request uses the legacy filter shape", func(t *testing.T) {
		req := BuildRequest(AirlineQueryID, "AIRLINE", "fr", 8728979, 0, 1)
		assert.Len(t, req, 1)
		v := req[0].Variables.(AirlineVariables)
		assert.Equal(t, Filters{{Axis: "LANGUAGE", Selections: []string{"fr"}}}, v.Filters)
	})
}

func TestExtractMichelinInfo(t *testing.T) {
	tests := []struct {
		name      string
		responses *Responses
		expected  *MichelinInfo
	}{
		{
			name:      "nil responses returns nil",
			responses: nil,
			expected:  nil,
		},
		{
			name:      "empty responses returns nil",
			responses: &Responses{},
			expected:  nil,
		},
		{
			name: "extracts Michelin info with awards",
			responses: &Responses{{Data: ResponseData{Michelin: []MichelinInfo{{
				AwardHeader: "MICHELIN Guide",
				Awards:      []MichelinAward{{AwardName: "1 Star", YearOfAward: "2024"}},
			}}}}},
			expected: &MichelinInfo{
				AwardHeader: "MICHELIN Guide",
				Awards: []MichelinAward{
					{AwardName: "1 Star", YearOfAward: "2024"},
				},
			},
		},
		{
			name:      "returns nil when Michelin data has no header and no awards",
			responses: &Responses{{Data: ResponseData{Michelin: []MichelinInfo{{}}}}},
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMichelinInfo(tt.responses)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortReviewsByDate(t *testing.T) {
	tests := []struct {
		name        string
		reviews     []Review
		expectedIDs []int
	}{
		{
			name:        "nil reviews does not panic",
			reviews:     nil,
			expectedIDs: nil,
		},
		{
			name:        "empty reviews does not panic",
			reviews:     []Review{},
			expectedIDs: []int{},
		},
		{
			name: "sorts newest first",
			reviews: []Review{
				{ID: 1, CreatedDate: "2024-01-01"},
				{ID: 2, CreatedDate: "2025-06-15"},
				{ID: 3, CreatedDate: "2024-07-20"},
			},
			expectedIDs: []int{2, 3, 1},
		},
		{
			name: "already sorted stays the same",
			reviews: []Review{
				{ID: 1, CreatedDate: "2025-12-01"},
				{ID: 2, CreatedDate: "2025-06-01"},
				{ID: 3, CreatedDate: "2025-01-01"},
			},
			expectedIDs: []int{1, 2, 3},
		},
		{
			name: "single review",
			reviews: []Review{
				{ID: 1, CreatedDate: "2025-01-01"},
			},
			expectedIDs: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortReviewsByDate(tt.reviews)
			if tt.expectedIDs == nil {
				assert.Nil(t, tt.reviews)
				return
			}
			ids := make([]int, len(tt.reviews))
			for i, r := range tt.reviews {
				ids[i] = r.ID
			}
			assert.Equal(t, tt.expectedIDs, ids)
		})
	}
}
