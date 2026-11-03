# syntax=docker/dockerfile:1
# The provisioner is a static binary built outside the image (CGO_ENABLED=0, see the CI workflow),
# so the runtime image only needs CA certificates, which distroless/static includes
FROM gcr.io/distroless/static-debian12:latest@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2

LABEL org.opencontainers.image.source="https://github.com/algo7/TripAdvisor-Review-Scraper"

# The compose file mounts the R2 credentials into /go/src/app/credentials, and exported CSV files are written here
WORKDIR /go/src/app

# Copy the views folder
COPY ./views ./views
COPY main .

# Open port 3000
EXPOSE 3000

# Runs as PID 1 without an init process: the provisioner handles SIGTERM itself and starts no child processes
CMD ["./main"]
