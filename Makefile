.PHONY: build run test clean docker-build docker-run fmt vet

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%s)
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

# Build the application
build:
	CGO_ENABLED=0 go build -o bot ./cmd/bot

# Build with version info
build-release:
	@echo "Building version $(VERSION) at timestamp $(BUILD_TIME)"
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bot ./cmd/bot

# Run the application
run:
	go run ./cmd/bot

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f bot
	rm -f coverage.out coverage.html
	rm -rf data/*.db*

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run linter
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t telegram-bot-go .

# Run Docker container
docker-run:
	docker run -d \
		--name telegram-bot \
		-p 8080:8080 \
		--env-file .env \
		-v $(PWD)/data:/root/data \
		telegram-bot-go

# Stop Docker container
docker-stop:
	docker stop telegram-bot
	docker rm telegram-bot

# View Docker logs
docker-logs:
	docker logs -f telegram-bot

# Install dependencies
deps:
	go mod download
	go mod tidy

# Create data directory
init:
	mkdir -p data

# Run all checks
check: fmt vet test

# Build for multiple platforms
build-all:
	@echo "Building version $(VERSION) for multiple platforms..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bot-linux-amd64 ./cmd/bot
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bot-linux-arm64 ./cmd/bot
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bot-darwin-amd64 ./cmd/bot
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bot-darwin-arm64 ./cmd/bot
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bot-windows-amd64.exe ./cmd/bot
