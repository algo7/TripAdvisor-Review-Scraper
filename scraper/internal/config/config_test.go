package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expected    *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "all defaults with only LOCATION_URL set",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				Languages:   []string{"en"},
				FileType:    "csv",
				ProxyHost:   "",
			},
		},
		{
			name:        "missing LOCATION_URL returns error",
			envVars:     map[string]string{},
			expectError: true,
			errorMsg:    "LOCATION_URL not set",
		},
		{
			name: "custom languages are parsed correctly",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				"LANGUAGES":    "en|fr|de",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				Languages:   []string{"en", "fr", "de"},
				FileType:    "csv",
				ProxyHost:   "",
			},
		},
		{
			name: "single custom language",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				"LANGUAGES":    "fr",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				Languages:   []string{"fr"},
				FileType:    "csv",
				ProxyHost:   "",
			},
		},
		{
			name: "FILETYPE json is accepted",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				"FILETYPE":     "json",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				Languages:   []string{"en"},
				FileType:    "json",
				ProxyHost:   "",
			},
		},
		{
			name: "FILETYPE csv is accepted",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				"FILETYPE":     "csv",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				Languages:   []string{"en"},
				FileType:    "csv",
				ProxyHost:   "",
			},
		},
		{
			name: "FILETYPE is case insensitive",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				"FILETYPE":     "JSON",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				Languages:   []string{"en"},
				FileType:    "json",
				ProxyHost:   "",
			},
		},
		{
			name: "invalid FILETYPE returns error",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				"FILETYPE":     "xml",
			},
			expectError: true,
			errorMsg:    "invalid file type",
		},
		{
			name: "PROXY_HOST is passed through",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				"PROXY_HOST":   "http://proxy:8080",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Test.html",
				Languages:   []string{"en"},
				FileType:    "csv",
				ProxyHost:   "http://proxy:8080",
			},
		},
		{
			name: "all env vars set",
			envVars: map[string]string{
				"LOCATION_URL": "https://www.tripadvisor.com/Airline_Review-d8729113-Reviews-Lufthansa",
				"LANGUAGES":    "en|fr|de|es|pt",
				"FILETYPE":     "json",
				"PROXY_HOST":   "socks5://proxy:1080",
			},
			expected: &Config{
				LocationURL: "https://www.tripadvisor.com/Airline_Review-d8729113-Reviews-Lufthansa",
				Languages:   []string{"en", "fr", "de", "es", "pt"},
				FileType:    "json",
				ProxyHost:   "socks5://proxy:1080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Setenv automatically restores the original value after the test
			// and unsets vars that were not previously set
			for _, key := range []string{"LOCATION_URL", "LANGUAGES", "FILETYPE", "PROXY_HOST"} {
				t.Setenv(key, "")
			}
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := NewConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
