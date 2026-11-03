FROM gcr.io/distroless/static-debian12:latest@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2

WORKDIR /

# The application has to be built first with CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -v -ldflags="-s -w" main.go outside of docker
COPY main .

CMD ["./main"]
