.PHONY: build build-stripped build-all clean deps test

VERSION ?= dev

# Build regular binary
build:
	go build -o finks cmd/finks/main.go

# Build stripped binary (smaller size, no debug info)
build-stripped:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o finks cmd/finks/main.go

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o finks-linux-amd64 ./cmd/finks/main.go
	@GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o finks-linux-arm64 ./cmd/finks/main.go
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o finks-darwin-amd64 ./cmd/finks/main.go
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o finks-darwin-arm64 ./cmd/finks/main.go
	@chmod +x finks-*

# Clean build artifacts
clean:
	rm -f finks finks-*

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./...

