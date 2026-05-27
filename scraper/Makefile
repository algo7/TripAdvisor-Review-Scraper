# Variables
NAME:= scraper
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

##@ General
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: test
test: fmt vet ## Run the tests
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html


##@ Build
bin:
	@mkdir -p bin

build: bin ## Build the application only
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath -v -ldflags="-s -w" -o bin/${NAME}-${GOOS}-${GOARCH} main.go

build-ci: ## Build the application for CI
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -v -ldflags="-s -w" -o main main.go

run: build ## Start the application in foreground
	./bin/${NAME}-${GOOS}-${GOARCH}

clean: ## Clean the build artifacts
	rm -rf bin
