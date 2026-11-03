# syntax=docker/dockerfile:1
# Base image
FROM golang:1.27.3-alpine3.24 AS base

LABEL org.opencontainers.image.source="https://github.com/algo7/TripAdvisor-Review-Scraper"

WORKDIR /go/src/app

# Install dumb-init to run as PID 1, forward signals to the provisioner and reap zombie processes
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
