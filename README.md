# MQTT-AI Shell Executor

A Dockerized Go application that listens to MQTT messages and executes AI-generated shell commands.

## Features

- Connects to MQTT broker (`mqtts://queue-dev.esper.cloud:443`)
- Listens for messages with format: `{"prompt": "your command prompt"}`
- Calls local AI model API to generate shell scripts
- Executes the generated shell scripts and prints output
- Fully containerized with Docker

## Configuration

The application can be configured using environment variables:

- `MQTT_BROKER`: MQTT broker URL (default: `mqtts://queue-dev.esper.cloud:443`)
- `MQTT_USERNAME`: MQTT username for authentication (optional)
- `MQTT_PASSWORD`: MQTT password for authentication (optional)
- `AI_MODEL_URL`: AI model API URL (default: `http://localhost:8080/generate-shell`)
- `MQTT_TOPIC`: MQTT topic to subscribe to (default: `commands`)
- `MQTT_CLIENT_ID`: MQTT client ID (default: `go-mqtt-ai-executor`)

## Build and Run

### Using Docker Compose (Recommended)

```bash
# Build and run the application
docker-compose up --build

# Run in detached mode
docker-compose up -d --build
```

### Using Docker

```bash
# Build the Docker image
docker build -t mqtt-ai-executor .

# Run the container
docker run --rm --name mqtt-ai-executor \
  --network host \
  -e MQTT_BROKER=mqtts://queue-dev.esper.cloud:443 \
  -e MQTT_USERNAME=your_username \
  -e MQTT_PASSWORD=your_password \
  -e AI_MODEL_URL=http://localhost:8080/generate-shell \
  -e MQTT_TOPIC=commands \
  mqtt-ai-executor
```

### Local Development

```bash
# Install dependencies
go mod tidy

# Run the application
go run .
```

## Usage

1. Make sure your AI model is running on `http://localhost:8080/generate-shell`
2. Start the MQTT-AI Shell Executor
3. Publish a message to the configured MQTT topic with the format:
   ```json
   {"prompt": "find all Python files in current directory"}
   ```
4. The application will:
   - Receive the MQTT message
   - Call the AI model API with the prompt
   - Execute the generated shell script
   - Print the output

## API Flow

1. **MQTT Message**: `{"prompt": "your command"}`
2. **AI Model Request**: 
   ```bash
   curl -X POST http://localhost:8080/generate-shell \
     -H "Content-Type: application/json" \
     -d '{"prompt": "your command"}'
   ```
3. **AI Model Response**: `{"shell_script": "generated script"}`
4. **Script Execution**: The shell script is executed using `/bin/sh -c`
5. **Output**: The script output is printed to stdout

## Security Considerations

- The application executes shell commands received from MQTT messages
- Ensure your MQTT broker is properly secured
- Consider implementing command validation/sanitization
- Run in a sandboxed environment for production use

## Logs

The application provides detailed logging for:
- MQTT connection status
- Received messages
- AI model API calls
- Shell script execution
- Errors and debugging information

## Dependencies

- `github.com/eclipse/paho.mqtt.golang`: MQTT client library
- Standard Go libraries for HTTP, JSON, and process execution

## Docker Image

The Docker image is built using multi-stage build:
- Build stage: Uses `golang:1.21-alpine` to compile the application
- Runtime stage: Uses `alpine:latest` with minimal dependencies
- Includes `ca-certificates` for HTTPS connections
- Includes `bash` for shell script execution
