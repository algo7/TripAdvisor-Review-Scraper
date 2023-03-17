# Base image
FROM golang:1.19 as build

WORKDIR /go/src/app
COPY . .

# Download packages
RUN go mod tidy

# Compile
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Multi-stage build using distroless image
FROM gcr.io/distroless/static-debian11
LABEL "org.opencontainers.image.source" = "https://github.com/algo7/TripAdvisor-Review-Scraper"

# Copy the binary
COPY --from=BUILD /go/bin/app ./

# Copy the rest
COPY . .

# Start the app
CMD ["./app"]