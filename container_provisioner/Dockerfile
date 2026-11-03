# syntax=docker/dockerfile:1
# Base image
FROM golang:1.27.3-alpine3.24 AS base

LABEL org.opencontainers.image.source="https://github.com/algo7/TripAdvisor-Review-Scraper"

WORKDIR /go/src/app

# Copy the views folder
COPY ./views ./views
COPY main .

# Open port 3000
EXPOSE 3000

# Runs as PID 1 without an init process: the provisioner handles SIGTERM itself and starts no child processes
CMD ["./main"]
