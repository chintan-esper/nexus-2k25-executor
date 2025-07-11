package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTMessage represents the incoming MQTT message structure
type MQTTMessage struct {
	Prompt string `json:"prompt"`
}

// AIResponse represents the response from the AI model
type AIResponse struct {
	ShellScript string `json:"shell_script"`
}

// AIRequest represents the request to the AI model
type AIRequest struct {
	Prompt string `json:"prompt"`
}

const (
	defaultMQTTBroker = "mqtts://queue-dev.esper.cloud:443"
	defaultAIModelURL = "http://localhost:8080/generate-shell"
	defaultMQTTTopic  = "hack-2k25-commands"
)

var config *Config

type Config struct {
	MQTTBroker   string
	MQTTUsername string
	MQTTPassword string
	AIModelURL   string
	MQTTTopic    string
	MQTTClientID string
}

func loadConfig() *Config {
	config := &Config{
		MQTTBroker:   "mqtts://qmqtt-url:443",
		MQTTUsername: "username",
		MQTTPassword: "password",
		AIModelURL:   "http://localhost:8080/generate-shell",
		MQTTTopic:    "hack-2k25-commands",
		MQTTClientID: "go-mqtt-ai-executor",
	}

	log.Printf("Configuration loaded:")
	log.Printf("  MQTT Broker: %s", config.MQTTBroker)
	log.Printf("  AI Model URL: %s", config.AIModelURL)
	log.Printf("  MQTT Topic: %s", config.MQTTTopic)
	log.Printf("  MQTT Client ID: %s", config.MQTTClientID)

	return config
}

func main() {
	log.Println("Starting MQTT-AI Shell Executor...")

	// Load configuration
	config = loadConfig()

	// Create MQTT client
	client := createMQTTClient()
	if client == nil {
		log.Fatal("Failed to create MQTT client")
	}

	// Connect to MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Failed to connect to MQTT broker:", token.Error())
	}
	log.Println("Connected to MQTT broker")

	// Subscribe to topic
	if token := client.Subscribe(config.MQTTTopic, 1, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatal("Failed to subscribe to topic:", token.Error())
	}
	log.Printf("Subscribed to topic: %s", config.MQTTTopic)

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Keep the application running until interrupted
	log.Println("Application is running. Press Ctrl+C to stop...")
	<-signalChan

	// Graceful shutdown
	log.Println("Shutting down gracefully...")

	// Unsubscribe from topic
	if token := client.Unsubscribe(config.MQTTTopic); token.Wait() && token.Error() != nil {
		log.Printf("Error unsubscribing from topic: %v", token.Error())
	} else {
		log.Printf("Unsubscribed from topic: %s", config.MQTTTopic)
	}

	// Disconnect from MQTT broker
	client.Disconnect(250)
	log.Println("Disconnected from MQTT broker")

	log.Println("Application stopped successfully")
}

func createMQTTClient() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MQTTBroker)
	opts.SetClientID(config.MQTTClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(time.Second * 5)
	opts.SetKeepAlive(time.Second * 30)

	// Set username and password if provided
	if config.MQTTUsername != "" {
		opts.SetUsername(config.MQTTUsername)
	}
	if config.MQTTPassword != "" {
		opts.SetPassword(config.MQTTPassword)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	opts.SetTLSConfig(tlsConfig)

	// Set connection lost handler
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	// Set reconnect handler
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
	})

	return mqtt.NewClient(opts)
}

func messageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))

	// Parse the incoming message
	var mqttMsg MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("Error parsing MQTT message: %v", err)
		return
	}

	if mqttMsg.Prompt == "" {
		log.Println("Empty prompt received, ignoring...")
		return
	}

	// Call AI model
	shellScript, err := callAIModel(mqttMsg.Prompt)
	if err != nil {
		log.Printf("Error calling AI model: %v", err)
		return
	}

	// Execute shell script
	if err := executeShellScript(shellScript); err != nil {
		log.Printf("Error executing shell script: %v", err)
		return
	}
}

func callAIModel(prompt string) (string, error) {
	log.Printf("Calling AI model with prompt: %s", prompt)

	// Prepare request
	aiReq := AIRequest{Prompt: prompt}
	jsonData, err := json.Marshal(aiReq)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// Make HTTP request
	resp, err := http.Post(config.AIModelURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI model returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse AI response
	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", fmt.Errorf("error parsing AI response: %v", err)
	}

	log.Printf("AI model returned shell script -=-=-=-=-=-=-=-=-=-==--: %s", aiResp.ShellScript)
	return aiResp.ShellScript, nil
}

func executeShellScript(script string) error {
	log.Printf("Executing shell script: %s", script)

	// Execute the shell script
	cmd := exec.Command("/bin/sh", "-c", script)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	// Print the output regardless of success or failure
	fmt.Printf("Shell script output:\n%s\n", string(output))

	if err != nil {
		log.Printf("Shell script execution failed: %v", err)
		return err
	}

	log.Println("Shell script executed successfully")
	return nil
}
