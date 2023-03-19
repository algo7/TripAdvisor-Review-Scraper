# syntax=docker/dockerfile:1
# Base image
FROM golang:1.20.2-alpine3.17 as base

LABEL "org.opencontainers.image.source" = "https://github.com/algo7/TripAdvisor-Review-Scraper"

WORKDIR /go/src/app

# Install dumb-init to prevent go fiber from exiting
# https://github.com/gofiber/fiber/issues/1036#issuecomment-841763449
RUN apk add dumb-init

# Copy the views folder
COPY ./views ./views

# Stage 2
FROM base as build

# For caching
COPY go.mod go.sum ./

# Download packages
RUN apk add upx && go mod tidy

# Copy the rest
COPY . .

# Compile
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /go/bin/app_orig


# Compress
RUN upx --best -o /go/bin/app /go/bin/app_orig


# Stage 3
FROM base as release

# Copy the binary
COPY --from=build /go/bin/app ./

# Open port 3000
EXPOSE 3000

# Start the app
# Has to be started with a shell, otherwise Fiber throws an err
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./app"]