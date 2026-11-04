package utils

import "testing"

const (
	hotelURL      = "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne_Canton_of_Vaud.html"
	restaurantURL = "https://www.tripadvisor.com/Restaurant_Review-g187265-d11827759-Reviews-La_Terrasse-Lyon_Rhone_Auvergne_Rhone_Alpes.html"
	airlineURL    = "https://www.tripadvisor.com/Airline_Review-d8729113-Reviews-Lufthansa"
	attractionURL = "https://www.tripadvisor.com/Attraction_Review-g187261-d1008501-Reviews-Les_Ailes_du_Mont_Blanc-Chamonix_Haute_Savoie_Auvergne_Rhone_Alpes.html"
)

func TestDetectScrapeType(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{url: hotelURL, want: "HOTEL"},
		{url: "https://www.tripadvisor.com/Hotel_Review-g60763-d93450-Reviews-The_Plaza-New_York_City_New_York.html", want: "HOTEL"},
		{url: restaurantURL, want: "RESTO"},
		{url: airlineURL, want: "AIRLINE"},
		{url: attractionURL, want: "ATTRACTION"},
		{url: "https://www.tripadvisor.fr/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne_Canton_of_Vaud.html", want: ""},
		{url: "https://www.tripadvisor.com/Search?q=lausanne", want: ""},
		{url: "https://www.tripadvisor.com/Hotels-g188107-Lausanne_Canton_of_Vaud-Hotels.html", want: ""},
		{url: "Beau Rivage Palace", want: ""},
		{url: "", want: ""},
	}

	for _, tt := range tests {
		got := DetectScrapeType(tt.url)
		if got != tt.want {
			t.Errorf("DetectScrapeType(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestCleanTripAdvisorURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{url: hotelURL, want: hotelURL},
		{url: "  " + hotelURL + "\n", want: hotelURL},
		{url: hotelURL + "#REVIEWS", want: hotelURL},
		{url: airlineURL + "?filterLang=en#REVIEWS", want: airlineURL},
	}

	for _, tt := range tests {
		got := CleanTripAdvisorURL(tt.url)
		if got != tt.want {
			t.Errorf("CleanTripAdvisorURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestIsTripAdvisorCountrySite(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{url: hotelURL, want: false},
		{url: "https://www.tripadvisor.fr/Hotel_Review-g1-d2-Reviews-X-Y.html", want: true},
		{url: "https://www.tripadvisor.co.uk/Hotel_Review-g1-d2-Reviews-X-Y.html", want: true},
		{url: "https://www.tripadvisor.com.au/Hotel_Review-g1-d2-Reviews-X-Y.html", want: true},
		{url: "https://example.com/page", want: false},
	}

	for _, tt := range tests {
		got := IsTripAdvisorCountrySite(tt.url)
		if got != tt.want {
			t.Errorf("IsTripAdvisorCountrySite(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestGetLocationNameFromURL(t *testing.T) {
	tests := []struct {
		url        string
		scrapeType string
		want       string
	}{
		{url: hotelURL, scrapeType: "HOTEL", want: "Beau_Rivage_Palace"},
		{url: restaurantURL, scrapeType: "RESTO", want: "La_Terrasse"},
		{url: airlineURL, scrapeType: "AIRLINE", want: "Lufthansa"},
		{url: attractionURL, scrapeType: "ATTRACTION", want: "Les_Ailes_du_Mont_Blanc"},
	}

	for _, tt := range tests {
		got := GetLocationNameFromURL(tt.url, tt.scrapeType)
		if got != tt.want {
			t.Errorf("GetLocationNameFromURL(%q, %q) = %q, want %q", tt.url, tt.scrapeType, got, tt.want)
		}
	}
}
