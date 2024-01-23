# syntax=docker/dockerfile:1
# Base image
FROM golang:1.21.1-alpine3.18 as base

LABEL "org.opencontainers.image.source" = "https://github.com/algo7/TripAdvisor-Review-Scraper"

WORKDIR /go/src/app

# Install dumb-init to prevent go fiber from exiting
# https://github.com/gofiber/fiber/issues/1036#issuecomment-841763449
RUN apk add dumb-init

# Copy the views folder
COPY ./views ./views
COPY main .

# Open port 3000
EXPOSE 3000

# Start the app
# Has to be started with a shell, otherwise Fiber throws an err
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./main"]