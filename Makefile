.PHONY: all build linux windows macos clean install

BINARY_NAME=kern
# УДАЛИТЬ строку: VERSION=1.2.3 - версия берется из main.go

# Добавить платформо-специфичные флаги
BUILD_FLAGS_LINUX = -tags linux
BUILD_FLAGS_WINDOWS = -tags windows
BUILD_FLAGS_DARWIN = -tags darwin

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
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-macos-amd64 ./cmd/kern
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-macos-arm64 ./cmd/kern

# Целевые платформы
build-linux:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS_LINUX) -o bin/kern-linux ./cmd/kern

build-windows:
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS_WINDOWS) -o bin/kern.exe ./cmd/kern

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS_DARWIN) -o bin/kern-darwin-amd64 ./cmd/kern
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS_DARWIN) -o bin/kern-darwin-arm64 ./cmd/kern

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
	docker build -t kern:$(shell grep 'const version' cmd/kern/main.go | awk -F'"' '{print $$2}') .