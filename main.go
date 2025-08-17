package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Logger middleware that logs everything about the request
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
		}
		// Restore the body for the next handler
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// Log extensive request information
		log.Println("=== INCOMING REQUEST ===")
		log.Printf("Timestamp: %s", start.Format(time.RFC3339))
		log.Printf("Method: %s", r.Method)
		log.Printf("URL: %s", r.URL.String())
		log.Printf("Path: %s", r.URL.Path)
		log.Printf("Raw Query: %s", r.URL.RawQuery)
		log.Printf("Protocol: %s", r.Proto)
		log.Printf("Host: %s", r.Host)
		log.Printf("Remote Address: %s", r.RemoteAddr)
		log.Printf("Request URI: %s", r.RequestURI)
		log.Printf("Content Length: %d", r.ContentLength)
		log.Printf("Transfer Encoding: %v", r.TransferEncoding)
		log.Printf("Close: %t", r.Close)

		// Log all headers
		log.Println("--- HEADERS ---")
		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("Header: %s = %s", name, value)
			}
		}

		// Log query parameters
		if len(r.URL.Query()) > 0 {
			log.Println("--- QUERY PARAMETERS ---")
			for key, values := range r.URL.Query() {
				for _, value := range values {
					log.Printf("Query Param: %s = %s", key, value)
				}
			}
		}

		// Log request body if present
		if len(body) > 0 {
			log.Println("--- REQUEST BODY ---")
			log.Printf("Body Length: %d bytes", len(body))
			log.Printf("Body Content: %s", string(body))
		}

		// Log form data if present
		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			if err := r.ParseForm(); err == nil {
				log.Println("--- FORM DATA ---")
				for key, values := range r.Form {
					for _, value := range values {
						log.Printf("Form Field: %s = %s", key, value)
					}
				}
			}
		}

		// Create a response writer wrapper to capture response details
		responseWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     200,
			responseBody:   make([]byte, 0),
		}

		// Call the next handler
		next.ServeHTTP(responseWriter, r)

		// Log response information
		duration := time.Since(start)
		log.Println("--- RESPONSE ---")
		log.Printf("Status Code: %d", responseWriter.statusCode)
		log.Printf("Response Body Length: %d bytes", len(responseWriter.responseBody))
		log.Printf("Response Body: %s", string(responseWriter.responseBody))
		log.Printf("Duration: %v", duration)
		log.Println("=== END REQUEST ===")
		log.Println()
	})
}

// Response writer wrapper to capture response data
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	responseBody []byte
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	rw.responseBody = append(rw.responseBody, b...)
	return rw.ResponseWriter.Write(b)
}

// Generic handler that serves static JSON responses
// This handler accepts ANY JSON payload without validation
func serveStaticJSON(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Served-By", "dummy-logger-server")
		w.Header().Set("X-Timestamp", time.Now().Format(time.RFC3339))

		// For POST/PUT requests, try to parse and log the JSON payload (but don't validate structure)
		if r.Method == "POST" || r.Method == "PUT" {
			// Re-read the body since it was already read in the logging middleware
			body, err := io.ReadAll(r.Body)
			if err == nil && len(body) > 0 {
				// Try to parse as JSON to validate it's valid JSON (but accept any structure)
				var jsonPayload interface{}
				if err := json.Unmarshal(body, &jsonPayload); err != nil {
					log.Printf("⚠️  WARNING: Invalid JSON payload received: %v", err)
					log.Printf("Raw payload: %s", string(body))
				} else {
					log.Printf("✅ Valid JSON payload received and logged above")
					// Pretty print the parsed JSON
					prettyJSON, _ := json.MarshalIndent(jsonPayload, "", "  ")
					log.Printf("Parsed JSON structure:\n%s", string(prettyJSON))
				}
			}
		}

		// Read the static JSON file
		filePath := filepath.Join("responses", filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file %s: %v", filePath, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to read response file",
				"file":  filename,
			})
			return
		}

		// For POST requests, return 201 Created
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
		}

		// Write the static JSON response (completely independent of input payload)
		w.Write(data)
	}
}

// Handler for DELETE requests
func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Served-By", "dummy-logger-server")
	w.Header().Set("X-Timestamp", time.Now().Format(time.RFC3339))
	w.WriteHeader(http.StatusNoContent)
}

// Catch-all handler for unmatched routes, but dont return error
func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("!!! UNMATCHED ROUTE !!! Method: %s, Path: %s", r.Method, r.URL.Path)
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Served-By", "dummy-logger-server")
	w.Header().Set("X-Timestamp", time.Now().Format(time.RFC3339))
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"method":  r.Method,
		"path":    r.URL.Path,
		"message": "This route is not defined in the OpenAPI specification",
		"body":    r.Body,
		"url":     r.URL,
		"header":  r.Header,
		"status":  http.StatusOK,
	}
	
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Create router
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Define routes based on OpenAPI specification
	// Users endpoints
	r.HandleFunc("/users", serveStaticJSON("users.json")).Methods("GET")
	r.HandleFunc("/users", serveStaticJSON("user.json")).Methods("POST")
	r.HandleFunc("/users/{id}", serveStaticJSON("user.json")).Methods("GET")
	r.HandleFunc("/users/{id}", serveStaticJSON("user.json")).Methods("PUT")
	r.HandleFunc("/users/{id}", handleDelete).Methods("DELETE")

	// Products endpoints
	r.HandleFunc("/products", serveStaticJSON("products.json")).Methods("GET")
	r.HandleFunc("/products/{id}", serveStaticJSON("product.json")).Methods("GET")

	// Orders endpoints
	r.HandleFunc("/orders", serveStaticJSON("order.json")).Methods("POST")

	// Health check endpoint (not in OpenAPI but useful)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"server":    "dummy-logger-go-server",
		})
	}).Methods("GET")

	// Echo endpoint - returns only the request body
	r.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body for echo: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		// If body is JSON, try to parse and return as JSON
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") && len(body) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(body, &jsonBody); err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(jsonBody)
				return
			}
		}

		// For non-JSON or invalid JSON, return the raw body
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	// Catch-all handler for unmatched routes (must be last)
	r.PathPrefix("/").HandlerFunc(catchAllHandler)

	// Start server
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("Starting dummy logger server on port %s", port)
	log.Printf("Server will log all incoming requests extensively")
	log.Printf("Static JSON responses are served from the 'responses' directory")
	log.Println("Available endpoints:")
	log.Println("  GET    /users")
	log.Println("  POST   /users")
	log.Println("  GET    /users/{id}")
	log.Println("  PUT    /users/{id}")
	log.Println("  DELETE /users/{id}")
	log.Println("  GET    /products")
	log.Println("  GET    /products/{id}")
	log.Println("  POST   /orders")
	log.Println("  GET    /health")
	log.Println("  *      /echo     (returns what it receives)")
	log.Println()

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
