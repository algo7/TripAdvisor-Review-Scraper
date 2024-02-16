package tripadvisor

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/graphql-go/graphql"
)

var photoSizeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PhotoSize",
	Fields: graphql.Fields{
		"width": &graphql.Field{
			Type: graphql.Int,
		},
		"height": &graphql.Field{
			Type: graphql.Int,
		},
		"url": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Define the photoType
var photoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Photo",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"statuses": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"photoSizes": &graphql.Field{
			Type: graphql.NewList(photoSizeType),
		},
	},
})

// Define the photosType
var photosType = graphql.NewList(photoType)

// Define the contributionCountsType
var contributionCountsType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ContributionCounts",
	Fields: graphql.Fields{
		"sumAllUgc": &graphql.Field{
			Type: graphql.Int,
		},
		"sumAllLikes": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

// Define the hometownType
var hometownType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Hometown",
	Fields: graphql.Fields{
		"locationId": &graphql.Field{
			Type: graphql.String,
		},
		"location": &graphql.Field{
			Type: graphql.String,
		},
		"fallbackString": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Define the routeType
var routeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Route",
	Fields: graphql.Fields{
		"url": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Define the avatarType
var avatarType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Avatar",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"photoSizes": &graphql.Field{
			Type: graphql.NewList(photoSizeType), // photoSizeType est défini précédemment
		},
	},
})

// Define the userProfileType
var userProfileType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UserProfile",
	Fields: graphql.Fields{
		"isMe": &graphql.Field{
			Type: graphql.Boolean,
		},
		"isVerified": &graphql.Field{
			Type: graphql.Boolean,
		},
		"contributionCounts": &graphql.Field{
			Type: contributionCountsType,
		},
		"isFollowing": &graphql.Field{
			Type: graphql.Boolean,
		},
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"userId": &graphql.Field{
			Type: graphql.String,
		},
		"displayName": &graphql.Field{
			Type: graphql.String,
		},
		"username": &graphql.Field{
			Type: graphql.String,
		},
		"hometown": &graphql.Field{
			Type: hometownType,
		},
		"route": &graphql.Field{
			Type: routeType,
		},
		"avatar": &graphql.Field{
			Type: avatarType,
		},
	},
})

// Define the reviewType
var reviewType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Review",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"createdDate": &graphql.Field{
			Type: graphql.String,
		},
		"publishedDate": &graphql.Field{
			Type: graphql.String,
		},
		"rating": &graphql.Field{
			Type: graphql.Int,
		},
		"publishPlatform": &graphql.Field{
			Type: graphql.String,
		},
		"tripInfo": &graphql.Field{
			Type: graphql.NewObject(graphql.ObjectConfig{
				Name: "TripInfo",
				Fields: graphql.Fields{
					"stayDate": &graphql.Field{
						Type: graphql.String,
					},
					"tripType": &graphql.Field{
						Type: graphql.String,
					},
				},
			}),
		},
		"photoIds": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
		"locationId": &graphql.Field{
			Type: graphql.Int,
		},
		"labels": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"text": &graphql.Field{
			Type: graphql.String,
		},
		"url": &graphql.Field{
			Type: graphql.String,
		},
		"photos": &graphql.Field{
			Type: photosType,
		},
		"userProfile": &graphql.Field{
			Type: userProfileType,
		},
		"username": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Review is a struct that represents the review object in the response body from TripAdvisor endpoints
func getReviews(source *os.File) ([]Review, error) {
	// Checks if source is open
	if source == nil {
		return nil, fmt.Errorf("source is nil")
	}
	// Decode the JSON data into a feedBack struct
	var feedBack Feedback
	err := json.NewDecoder(source).Decode(&feedBack)
	if err != nil {
		return nil, err
	}
	return feedBack.Reviews, nil
}

// CreateSchemaFromFile is a function that creates a new GraphQL schema from a source file
func CreateSchemaFromFile(source *os.File) (graphql.Schema, error) {
	reviews, err := getReviews(source)
	if err != nil {
		return graphql.Schema{}, err
	}
	// Define the fields for the root query
	return CreateSchemaFromLocalData(reviews)
}

// CreateSchemaFromLocalData is a function that creates a new GraphQL schema
func CreateSchemaFromLocalData(reviews []Review) (graphql.Schema, error) {
	// Define the fields for the root query
	fields := graphql.Fields{
		"reviews": &graphql.Field{
			Type: graphql.NewList(reviewType),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"ratingMin": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "Filter value for rating. If provided, only reviews with a rating greater than or equal to this value will be returned.",
				},
				"ratingMax": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "Filter value for rating. If provided, only reviews with a rating less than or equal to this value will be returned.",
				},
				"rating": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "Filter value for rating. If provided, only reviews with a rating equal to this value will be returned.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				idQuery, isOK := p.Args["id"].(int)
				if isOK {
					for _, review := range reviews {
						if review.ID == idQuery {
							return []Review{review}, nil
						}
					}
				}
				ratingMin, isOK := p.Args["ratingMin"].(int)
				if isOK {
					var filteredReviews []Review
					for _, review := range reviews {
						if review.Rating >= ratingMin {
							filteredReviews = append(filteredReviews, review)
						}
					}
					return filteredReviews, nil
				}
				ratingMax, isOK := p.Args["ratingMax"].(int)
				if isOK {
					var filteredReviews []Review
					for _, review := range reviews {
						if review.Rating < ratingMax {
							filteredReviews = append(filteredReviews, review)
						}
					}
					return filteredReviews, nil
				}
				rating, isOK := p.Args["rating"].(int)
				if isOK {
					var filteredReviews []Review
					for _, review := range reviews {
						if review.Rating == rating {
							filteredReviews = append(filteredReviews, review)
						}
					}
					return filteredReviews, nil
				}
				return reviews, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	return graphql.NewSchema(schemaConfig)
}
