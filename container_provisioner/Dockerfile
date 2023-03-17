# Base image
FROM golang:1.19 as build

WORKDIR /go/src/app
COPY . .

# Download packages
RUN go mod tidy

# Compile
RUN go build -o /go/bin/app

# Multi-stage build using distroless image
FROM gcr.io/distroless/static-debian11

# Copy the binary
COPY --from=build /go/bin/app /

# Start the app
CMD ["/app"]