# TripAdvisor Review Scraper

## Build Instructions

This scraper is written in Go and requires Go +v1.21 to build.

You can build the project in 2 ways:

1. Using Make
   - Simply run `make build` in the root directory of the project.
2. Using Go CLI
   - Run `go build main.go` in the root directory of the project.

### Note on Docker

The scraper can be run from a docker container. To build the container, you need to build the Go binary first using one of the 2 methods above and then build the container using the Dockerfile located in the root directory of the project.

After you have built the binary, you can build the container using the following command:

```bash
docker build -t <image_name>:<tag> .
```

## Usage Instructions

The scraper needs a single environment variable to run: `LOCATION_URL`, which is the URL of the TripAdvisor page of the hotel/restaurant/airline you want to scrape. The URL should be in the following format:

1. Airline: `https://www.tripadvisor.com/Airline_Review-d8729113-Reviews-Lufthansa`
2. Hotel: `https://www.tripadvisor.com/Hotel_Review-g188107-d231860-Reviews-Beau_Rivage_Palace-Lausanne_Canton_of_Vaud.html`
3. Restaurant: `https://www.tripadvisor.com/Restaurant_Review-g187265-d11827759-Reviews-La_Terrasse-Lyon_Rhone_Auvergne_Rhone_Alpes.html`
4. Attraction: `https://www.tripadvisor.com/Attraction_Review-g187261-d1008501-Reviews-Les_Ailes_du_Mont_Blanc-Chamonix_Haute_Savoie_Auvergne_Rhone_Alpes.html`

Note that the URL has to be from  `https://www.tripadvisor.com` and not other TripAdvisor domains such as `.fr`, `.ch`, `.de`, etc.

The scraper may use a `LANGUAGES` environment variable to specify the languages in which to scrape the reviews. The languages should be | and in the format `en|fr|de|es|pt`. If the `LANGUAGES` environment variable is not set, the scraper will default to English.

The scraper may use a `FILETYPE` environment variable to specify the file type in which to save the scraped data. The file type should be `json` or `csv`. If the `FILETYPE` environment variable is not set, the scraper will default to `csv`.  
The json filetype will be more verbose and will contain all the data scraped from the TripAdvisor page. The csv filetype will contain only the review text and the review rating.

Run using the binary directly:

```bash
export LOCATION_URL=<TripAdvisor_URL>
## optional
export LANGUAGES="en|fr|de|es|pt"
./binary_name
```

Run using Docker:

```bash
docker run -e LANGUAGES="en|fr|de|es|pt" LOCATION_URL=<TripAdvisor_URL> <image_name>:<tag>
```

## GraphQL API

An experimental way to expose the result is through a GraphQL API. The API exposes a single endpoint `/graphql` which accepts a single query `reviews`.
for example:

```bash
curl -X POST -H "Content-Type: application/json" --data '{ "query": "{ reviews { id rating } }" }' http://localhost:8080/graphql
curl -X POST -H "Content-Type: application/json" --data '{ "query": "{ reviews(rating:4) { id rating } }" }' http://localhost:8080/graphql
curl -X POST -H "Content-Type: application/json" --data '{ "query": "{ reviews(ratingMax:4) { id rating title text} }" }' http://localhost:8080/graphql
curl -X POST -H "Content-Type: application/json" --data '{ "query": "{ reviews(id:564320144) { id rating } }" }' http://localhost:8080/graphql
curl -X POST -H "Content-Type: application/json" --data '{ "query": "{ reviews(id: 822288866) { id rating title text} }" }' http://localhost:8080/graphql
```

For using this **experimental** feature, you need to set the `FILETYPE` environment variable to `graphql` and the `PORT` environment variable to the port you want the API to listen on. The default port is `8080`.
