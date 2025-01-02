# Simple  Notification Service

### Features

- Produce simple event use http request.
- Consume events from RabbitMQ.
- Broadcast events to WebSocket clients.
- WebSocket clients can subscribe to specific event types or unsubscribe from specific event.
- Ping-pong mechanism for health checks.
- Handle Dead Letter Queue for prevent lost message if clint not ready for give message.

## Setup Instructions

### Prerequisites

- Docker and Docker Compose installed.

### Run the Service

1. Build and run the Docker images:
   ```bash
   make build
   ```
2. Run the Docker images:
   ```bash
   make up
   ``` 
3. Stop the service:
   ```bash
   make down
   ```
4. Connect WebSocket clients to `ws://localhost:8080/ws`.

---

## How Communicate with Service ?

### 1. Produce an Event

This endpoint allows you to produce an event and send to RabbitMQ.

- **Endpoint**: `POST /event`
- **Request Body**:
  ```json
  {
      "event_type": "user_signup",
      "payload": {
          "name": "sajjad",
          "family": "rezaei"
      }
  }
  ```
- **Example Request**:
  ```bash
  curl -X POST http://localhost:8080/event \
  -H "Content-Type: application/json" \
  -d '{
      "event_type": "user_signup",
      "payload": {
          "name": "sajjad",
          "family": "rezaei"
      }
  }'
  ```
- **Response**:
    - `200 OK event published` if the event is successfully sent to RabbitMQ.
    - `400 Bad Request` if the request body is invalid.

  ```bash
  curl -X POST http://localhost:8080/event \
  -H "Content-Type: application/json" \
  -d '{
      "event_type": "order_created",
      "payload": {
          "name": "sajjad",
          "family": "rezaei"
      }
  }'
  ```
- **Response**:
    - `200 OK event published` if the event is successfully sent to RabbitMQ.
    - `400 Bad Request` if the request body is invalid.
---

### 1. RabbitMq panel `http://localhost:15672`

```
Username: guest
Password: guest
```

## WebSocket Client

### 1. Connect to WebSocket

Clients can connect to the WebSocket server to receive real-time notifications.

- **Endpoint**: `ws://localhost:8080/ws`

### 2. Send Messages over WebSocket

After connecting, clients can send JSON messages to perform actions like subscribing to events, unsubscribing, or
sending pings.

#### a. Subscribe to an Event

- **Request**:
  ```json
  {
      "action": "subscribe",
      "event_type": "user_signup"
  }
  ```
  This will subscribe the client to the `user_signup` events. The client will receive a notification whenever an event
  of this type is broadcast.

#### b. Unsubscribe from an Event

- **Request**:
  ```json
  {
      "action": "unsubscribe",
      "event_type": "user_signup"
  }
  ```
  This will unsubscribe the client from the `user_signup` events.

#### c. Ping

- **Request**:
  ```json
  {
      "action": "ping"
  }
  ```
  This is used to check the health of the WebSocket connection. The server will respond with a `pong` message.

---

## Example WebSocket Client Code

Here’s an example of how to use a WebSocket client to interact with the service.

### Using JavaScript (Node.js or Browser Or Postman Or ...)

```javascript
const WebSocket = require('ws'); // For Node.js, install using `npm install ws`

const socket = new WebSocket('ws://localhost:8080/ws');

// Connection opened
socket.onopen = () => {
    console.log('Connected to WebSocket server');

    // Subscribe to an event type
    socket.send(JSON.stringify({
        action: "subscribe",
        event_type: "user_signup"
    }));

    // Send a ping
    socket.send(JSON.stringify({
        action: "ping"
    }));
};

// Listen for messages
socket.onmessage = (event) => {
    console.log('Message from server:', event.data);
};

// Listen for errors
socket.onerror = (error) => {
    console.error('WebSocket error:', error);
};

// Connection closed
socket.onclose = () => {
    console.log('WebSocket connection closed');
};

// Unsubscribe from an event type after 10 seconds
setTimeout(() => {
    socket.send(JSON.stringify({
        action: "unsubscribe",
        event_type: "user_signup"
    }));
}, 10000);
```

---

### Flow Diagram

1. Clients connect via WebSocket (`ws://localhost:8080/ws`).
2. Clients subscribe to specific event types (e.g., `user_signup`,`order_created`).
3. Events are produced via the `/event` API and sent to RabbitMQ.
4. The service consumes events from RabbitMQ and broadcasts them to subscribed clients.

---

## Example Workflow

1. Start the service:
   ```bash
   make run
   ```
2. Produce an event:
   ```bash
   curl -X POST http://localhost:8080/event \
   -H "Content-Type: application/json" \
   -d '{
       "event_type": "user_signup",
       "payload": {
           "name": "sajjad",
           "family": "rezaei"
       }
   }'
   ```
3. Connect a WebSocket client to `ws://localhost:8080/ws`.
4. Subscribe to the `user_signup` event:
   ```json
   {
       "action": "subscribe",
       "event_type": "user_signup"
   }
   ```
5. The client will receive the event:
   ```json
   {
       "type": "user_signup",
       "payload": {
           "name": "sajjad",
           "family": "rezaei"
       }
   }
   ```
6. Unsubscribe from the `user_signup` event:
   ```json
   {
       "action": "unsubscribe",
       "event_type": "user_signup"
   }
   ```
7. Send a ping:
   ```json
   {
       "action": "ping"
   }
   ```

---

## Notes

- Ensure RabbitMQ is running on docker and exposed port 5672  `localhost:5672`. Update the connection string in the code
  if needed.
- The service is designed to handle multiple WebSocket connections efficiently using Goroutines.
- Use the provided Makefile for building and running the service.

---