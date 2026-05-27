package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRequestedByID(t *testing.T) {
	const expectedLength = 180
	const allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	tests := []struct {
		name string
	}{
		{name: "generates a valid ID"},
		{name: "generates a second valid ID"},
	}

	var previousID string

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := GenerateRequestedByID()
			assert.NoError(t, err)
			assert.Len(t, id, expectedLength)

			// Verify all characters are from the allowed set
			for i, c := range id {
				assert.Contains(t, allowedChars, string(c), "character at index %d is not in allowed set: %c", i, c)
			}

			// Verify uniqueness across calls
			if previousID != "" {
				assert.NotEqual(t, previousID, id, "two consecutive calls should produce different IDs")
			}
			previousID = id
		})
	}
}
