package tripadvisor

import (
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
				TripInfo: struct {
					StayDate string `json:"stayDate"`
					TripType string `json:"tripType"`
				}{
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
				TripInfo: struct {
					StayDate string `json:"stayDate"`
					TripType string `json:"tripType"`
				}{
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
				TripInfo: struct {
					StayDate string `json:"stayDate"`
					TripType string `json:"tripType"`
				}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReviewToCSVRow(tt.review, tt.locationName, tt.michelin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractReviews(t *testing.T) {
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
			name: "extracts from Locations path",
			responses: &Responses{
				{
					Data: struct {
						ReviewsProxy []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						} `json:"ReviewsProxy_getReviewListPageForLocation"`
						Locations []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						} `json:"locations"`
						Michelin []MichelinInfo `json:"RestaurantAwards_getRestaurantAwards"`
					}{
						Locations: []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						}{
							{
								LocationID: 12345,
								ReviewListPage: struct {
									TotalCount int      `json:"totalCount"`
									Reviews    []Review `json:"reviews"`
								}{
									Reviews: []Review{
										{ID: 1, Title: "Test Review"},
									},
								},
							},
						},
					},
				},
			},
			expected: []Review{{ID: 1, Title: "Test Review"}},
		},
		{
			name: "extracts from ReviewsProxy path",
			responses: &Responses{
				{
					Data: struct {
						ReviewsProxy []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						} `json:"ReviewsProxy_getReviewListPageForLocation"`
						Locations []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						} `json:"locations"`
						Michelin []MichelinInfo `json:"RestaurantAwards_getRestaurantAwards"`
					}{
						ReviewsProxy: []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						}{
							{
								Reviews: []Review{
									{ID: 100, Title: "Airline Review"},
								},
							},
						},
					},
				},
			},
			expected: []Review{{ID: 100, Title: "Airline Review"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractReviews(tt.responses)
			assert.Equal(t, tt.expected, result)
		})
	}
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
			name: "extracts count from Locations path",
			responses: &Responses{
				{
					Data: struct {
						ReviewsProxy []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						} `json:"ReviewsProxy_getReviewListPageForLocation"`
						Locations []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						} `json:"locations"`
						Michelin []MichelinInfo `json:"RestaurantAwards_getRestaurantAwards"`
					}{
						Locations: []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						}{
							{
								ReviewListPage: struct {
									TotalCount int      `json:"totalCount"`
									Reviews    []Review `json:"reviews"`
								}{
									TotalCount: 42,
								},
							},
						},
					},
				},
			},
			expected: 42,
		},
		{
			name: "extracts count from ReviewsProxy path",
			responses: &Responses{
				{
					Data: struct {
						ReviewsProxy []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						} `json:"ReviewsProxy_getReviewListPageForLocation"`
						Locations []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						} `json:"locations"`
						Michelin []MichelinInfo `json:"RestaurantAwards_getRestaurantAwards"`
					}{
						ReviewsProxy: []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						}{
							{TotalCount: 99},
						},
					},
				},
			},
			expected: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTotalCount(tt.responses)
			assert.Equal(t, tt.expected, result)
		})
	}
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
			responses: &Responses{
				{
					Data: struct {
						ReviewsProxy []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						} `json:"ReviewsProxy_getReviewListPageForLocation"`
						Locations []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						} `json:"locations"`
						Michelin []MichelinInfo `json:"RestaurantAwards_getRestaurantAwards"`
					}{
						Michelin: []MichelinInfo{
							{
								AwardHeader: "MICHELIN Guide",
								Awards: []MichelinAward{
									{AwardName: "1 Star", YearOfAward: "2024"},
								},
							},
						},
					},
				},
			},
			expected: &MichelinInfo{
				AwardHeader: "MICHELIN Guide",
				Awards: []MichelinAward{
					{AwardName: "1 Star", YearOfAward: "2024"},
				},
			},
		},
		{
			name: "returns nil when Michelin data has no header and no awards",
			responses: &Responses{
				{
					Data: struct {
						ReviewsProxy []struct {
							TotalCount        int      `json:"totalCount"`
							Reviews           []Review `json:"reviews"`
							ReviewListOptions struct {
								SortType string `json:"sortType"`
								SortBy   string `json:"sortBy"`
							} `json:"reviewListOptions"`
						} `json:"ReviewsProxy_getReviewListPageForLocation"`
						Locations []struct {
							LocationID     int `json:"locationId"`
							ReviewListPage struct {
								TotalCount int      `json:"totalCount"`
								Reviews    []Review `json:"reviews"`
							} `json:"reviewListPage"`
						} `json:"locations"`
						Michelin []MichelinInfo `json:"RestaurantAwards_getRestaurantAwards"`
					}{
						Michelin: []MichelinInfo{
							{},
						},
					},
				},
			},
			expected: nil,
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
