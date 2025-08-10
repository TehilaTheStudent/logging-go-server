# Dummy Logger Go Server

A comprehensive dummy Go server that logs all incoming requests extensively and serves static JSON responses based on an OpenAPI specification.

## Features

- **Extensive Logging**: Logs every aspect of incoming requests including:
  - Request method, URL, headers, query parameters
  - Request body content
  - Form data (if applicable)
  - Response status codes and body
  - Request duration
  - Remote address and protocol details

- **OpenAPI-Based Routes**: Server routes are defined based on the OpenAPI specification in `openapi.yaml`

- **Static JSON Responses**: All responses are served from static JSON files in the `responses/` directory

- **Catch-All Logging**: Logs requests that don't match any defined routes

## Directory Structure

```
dummy-logger-go-server/
├── main.go              # Main server implementation
├── go.mod               # Go module definition
├── openapi.yaml         # OpenAPI specification defining routes
├── responses/           # Static JSON response files
│   ├── users.json       # Response for GET /users
│   ├── user.json        # Response for user operations
│   ├── products.json    # Response for GET /products
│   ├── product.json     # Response for product operations
│   └── order.json       # Response for order operations
└── README.md            # This file
```

## Available Endpoints

Based on the OpenAPI specification:

- `GET /users` - Get all users
- `POST /users` - Create a new user
- `GET /users/{id}` - Get user by ID
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user
- `GET /products` - Get all products (supports query parameters)
- `GET /products/{id}` - Get product by ID
- `POST /orders` - Create a new order
- `GET /health` - Health check endpoint

## Running the Server

### Option 1: Run with Go (Local Development)

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Run the server:
   ```bash
   go run main.go
   ```

3. The server will start on port 8080 (or the port specified in the `PORT` environment variable)

### Option 2: Run with Docker

1. Build and run with Docker Compose (recommended):
   ```bash
   docker-compose up --build
   ```

2. Or build and run with Docker directly:
   ```bash
   # Build the image
   docker build -t dummy-logger-server .
   
   # Run the container
   docker run -p 8080:8080 dummy-logger-server
   ```

3. The containerized server will be available at http://localhost:8080

### Option 3: Build Binary

1. Build the executable:
   ```bash
   # On Windows
   build.bat
   
   # On Linux/Mac
   go build -o dummy-logger-server main.go
   ```

2. Run the binary:
   ```bash
   ./dummy-logger-server
   ```

## Example Usage

```bash
# Test with curl
curl -X GET http://localhost:8080/users
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name":"Test User","email":"test@example.com"}'
curl -X GET http://localhost:8080/products?category=electronics&limit=5
curl -X GET http://localhost:8080/nonexistent-route  # This will be logged as unmatched
```

## Logging Output

The server logs extensive information for each request:

```
=== INCOMING REQUEST ===
Timestamp: 2024-01-18T16:45:00Z
Method: POST
URL: /users
Path: /users
Protocol: HTTP/1.1
Host: localhost:8080
Remote Address: 127.0.0.1:54321
Content Length: 45
--- HEADERS ---
Header: Content-Type = application/json
Header: User-Agent = curl/7.68.0
--- REQUEST BODY ---
Body Length: 45 bytes
Body Content: {"name":"Test User","email":"test@example.com"}
--- RESPONSE ---
Status Code: 201
Response Body Length: 98 bytes
Response Body: {"id":"user-001","name":"John Doe","email":"john.doe@example.com","created_at":"2024-01-15T10:30:00Z"}
Duration: 2.345ms
=== END REQUEST ===
```

## Customization

- **Add new routes**: Update `openapi.yaml` and add corresponding handlers in `main.go`
- **Modify responses**: Edit JSON files in the `responses/` directory
- **Change logging format**: Modify the `loggingMiddleware` function in `main.go`
- **Add authentication**: Extend the middleware to log authentication headers

## Environment Variables

- `PORT`: Server port (default: 8080)

## Dependencies

- `github.com/gorilla/mux`: HTTP router for handling routes and middleware
