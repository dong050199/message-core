# Message Core

A multi-protocol messaging broker that supports MQTT and WebSocket protocols with Redis backend and Kafka integration.

## Overview

Message Core is a centralized messaging service that provides:

- MQTT broker for IoT device communication
- WebSocket server for real-time web applications
- Redis caching for message storage and session management
- Kafka integration for event streaming and message processing
- Custom hooks for access control and message transformation
- Platform service integration for user validation and rule application

## Features

- **MQTT Broker**: Built with [mochi-co/mqtt](https://github.com/mochi-co/mqtt) to handle MQTT connections
- **WebSocket Server**: Real-time bidirectional communication for web clients
- **Redis Integration**: Caching and pub/sub functionality
- **Kafka Support**: Reliable message processing and event streaming
- **Custom Hooks**: Authentication, authorization, and message transformation
- **Rule Engine**: Apply custom rules to messages based on topic and content

## Architecture

The service consists of the following components:

- **WebSocket Server**: Handles WebSocket connections on port 8080
- **MQTT Broker**: Handles MQTT connections on port 1883
- **Redis Client**: Connects to Redis for caching and pub/sub
- **Kafka Producer/Consumer**: Connects to Kafka for message processing

## Getting Started

### Prerequisites

- Go 1.20 or higher
- Docker and Docker Compose (for containerized deployment)
- Redis server
- Kafka broker

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd message-core
   ```

2. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

3. Build and run locally:
   ```bash
   go build -o main
   ./main
   ```

### Docker Deployment

Run the service with Docker Compose:

```bash
docker-compose up -d
```

This will start the Message Core service along with Redis and Grafana for monitoring.

## Configuration

The service is configured through environment variables. Key configuration options:

### Redis Configuration
- `REDIS_URL`: Redis server URL
- `REDIS_SIGLE_MODE`: Set to true for single Redis instance

### Kafka Configuration
- `KAFKA_BROKERS`: Comma-separated list of Kafka brokers
- `KAFKA_GROUP_ID`: Consumer group ID
- `KAFKA_TOPIC_*`: Various topic configurations

## Usage

### MQTT Client Connection

Connect MQTT clients to `localhost:1883` with appropriate credentials.

### WebSocket Client Connection

Connect WebSocket clients to `ws://localhost:8080/socket`.

### Message Format

Messages should follow the defined format:

```json
{
  "action": "<action>",
  "topic": "<topic>",
  "message": "<message-content>"
}
```

## Monitoring

Grafana is included in the Docker deployment for monitoring. Access it at http://localhost:3001 with:
- Username: admin
- Password: 1234abcd@@