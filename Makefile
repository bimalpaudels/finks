.PHONY: build build-stripped clean

# Build regular binary
build:
	go build -o finks cmd/finks/main.go

# Build stripped binary (smaller size, no debug info)
build-stripped:
	go build -ldflags="-s -w" -o finks cmd/finks/main.go


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

