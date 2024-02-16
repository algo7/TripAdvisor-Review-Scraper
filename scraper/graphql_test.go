package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/algo7/TripAdvisor-Review-Scraper/scraper/pkg/tripadvisor"
	"github.com/graphql-go/graphql"
)

var testFile = filepath.Join("reviews.json")

func TestCreateSchema(t *testing.T) {
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()
	schema, err := tripadvisor.CreateSchemaFromFile(file)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	if schema.QueryType() == nil {
		t.Errorf("CreateSchema() = %v, want non-nil", schema.QueryType())
	}

	// Test a query against the schema
	query := `
        {
            reviews {
                id
                createdDate
                publishedDate
                rating
                publishPlatform
            }
        }
    `
	params := graphql.Params{Schema: schema, RequestString: query}
	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if result.Data == nil {
		t.Errorf("CreateSchema() = %v, want non-nil", result.Data)
	}

	// Verify that the result is an array
	if data, ok := result.Data.(map[string]interface{}); ok {
		if reviews, ok := data["reviews"].([]interface{}); ok {
			// Verify that each item in the array has the required fields
			for _, review := range reviews {
				if reviewData, ok := review.(map[string]interface{}); ok {
					if _, ok := reviewData["id"]; !ok {
						t.Errorf("review missing id field")
					}
					if _, ok := reviewData["createdDate"]; !ok {
						t.Errorf("review missing createdDate field")
					}
					if _, ok := reviewData["publishedDate"]; !ok {
						t.Errorf("review missing publishedDate field")
					}
					if _, ok := reviewData["rating"]; !ok {
						t.Errorf("review missing rating field")
					}
					if _, ok := reviewData["publishPlatform"]; !ok {
						t.Errorf("review missing publishPlatform field")
					}
				} else {
					t.Errorf("review is not an object")
				}
			}
		} else {
			t.Errorf("reviews is not an array")
		}
	} else {
		t.Errorf("result.Data is not an object")
	}
}

func TestIdQuery(t *testing.T) {
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()
	schema, err := tripadvisor.CreateSchemaFromFile(file)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	if schema.QueryType() == nil {
		t.Errorf("CreateSchema() = %v, want non-nil", schema.QueryType())
	}

	// Test a query against the schema
	query := `
        {
            reviews(id: 822288866) {
                id
                title
				text
            }
        }
    `
	params := graphql.Params{Schema: schema, RequestString: query}
	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if result.Data == nil {
		t.Errorf("CreateSchema() = %v, want non-nil", result.Data)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Errorf("expected result.Data to be a map[string]interface{}")
		return
	}

	reviews, ok := data["reviews"].([]interface{})
	if !ok {
		t.Errorf("expected data[\"reviews\"] to be a []interface{}")
		return
	}

	if len(reviews) != 1 {
		t.Errorf("expected 1 result, got %d", len(reviews))
	}
	if review, ok := reviews[0].(map[string]interface{}); ok {
		if review["title"] != "Exceptionnel" {
			t.Errorf("expected title to be \"Exceptionnel\", got %v", review["title"])
		}
	} else {
		t.Errorf("expected reviews[0] to be a map[string]interface{}")
	}

}

func TestRatingMax(t *testing.T) {
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()
	schema, err := tripadvisor.CreateSchemaFromFile(file)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	if schema.QueryType() == nil {
		t.Errorf("CreateSchema() = %v, want non-nil", schema.QueryType())
	}

	// Test a query against the schema
	query := `
        {
            reviews(ratingMax: 2) {
                id
                title
				text
            }
        }
    `
	params := graphql.Params{Schema: schema, RequestString: query}
	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Errorf("expected result.Data to be a map[string]interface{}")
	}
	reviews, ok := data["reviews"].([]interface{})
	if !ok {
		t.Errorf("expected data[\"reviews\"] to be a []interface{}")
	}
	if len(reviews) != 0 {
		t.Errorf("expected 0 result, got %d", len(reviews))
	}
}

func TestRatingMin(t *testing.T) {
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()
	schema, err := tripadvisor.CreateSchemaFromFile(file)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	if schema.QueryType() == nil {
		t.Errorf("CreateSchema() = %v, want non-nil", schema.QueryType())
	}

	// Test a query against the schema
	query := `
        {
            reviews(ratingMin: 5) {
                id
                title
				text
            }
        }
    `
	params := graphql.Params{Schema: schema, RequestString: query}
	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Errorf("expected result.Data to be a map[string]interface{}")
	}
	reviews, ok := data["reviews"].([]interface{})
	if !ok {
		t.Errorf("expected data[\"reviews\"] to be a []interface{}")
	}
	if 135 < len(reviews) {
		t.Errorf("expected <135 result, got %d", len(reviews))
	}
}
