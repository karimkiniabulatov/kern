.PHONY: all build linux windows macos android clean install

BINARY_NAME=kern
VERSION=1.2.0

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) ./cmd/kern

linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux ./cmd/kern

windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME).exe ./cmd/kern
	GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o bin/$(BINARY_NAME)-service.exe ./cmd/kern

macos:
	@echo "Building for MacOS..."
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-macos ./cmd/kern

android:
	@echo "Building for Android..."
	cd android && ./scripts/setup-android.sh
	cd android && ./scripts/build-android.sh

release: linux windows macos
	@echo "Creating release packages..."
	./scripts/build-release.sh

install:
	@echo "Installing to /usr/local/bin..."
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	rm -rf bin/ dist/

test:
	@echo "Running tests..."
	go test ./internal/...

# Для разработки
dev:
	@echo "Starting in development mode..."
	go run ./cmd/kern --all --refresh=1

# Clean development build - always recreates dependencies
dev-clean:
	@echo "Clean development build..."
	rm -f go.sum
	go mod download
	go build -o bin/$(BINARY_NAME) ./cmd/kern

.PHONY: docker
docker:
	@echo "Building Docker image..."
	docker build -t kern:$(VERSION) .