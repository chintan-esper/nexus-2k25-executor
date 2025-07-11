.PHONY: build run docker-build docker-run clean test deps

# Default target
all: build

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build the application
build: deps
	go build -o bin/mqtt-ai-executor .

# Run the application locally
run: build
	./bin/mqtt-ai-executor

# Build Docker image
docker-build:
	docker build -t mqtt-ai-executor .

# Run Docker container
docker-run: docker-build
	docker run --rm --name mqtt-ai-executor \
		--network host \
		-e MQTT_BROKER=mqtts://queue-dev.esper.cloud:443 \
		-e AI_MODEL_URL=http://localhost:8080/generate-shell \
		-e MQTT_TOPIC=commands \
		mqtt-ai-executor

# Run with docker-compose
docker-compose-up:
	docker-compose up --build

# Run with docker-compose in detached mode
docker-compose-up-d:
	docker-compose up -d --build

# Stop docker-compose
docker-compose-down:
	docker-compose down

# Clean build artifacts
clean:
	rm -rf bin/
	docker rmi mqtt-ai-executor 2>/dev/null || true

# Test the application
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  run                - Run the application locally"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  docker-compose-up  - Run with docker-compose"
	@echo "  docker-compose-up-d - Run with docker-compose in detached mode"
	@echo "  docker-compose-down - Stop docker-compose"
	@echo "  clean              - Clean build artifacts"
	@echo "  test               - Run tests"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"
	@echo "  deps               - Install dependencies"
	@echo "  help               - Show this help message"
