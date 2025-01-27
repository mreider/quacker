// server.go

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// Flood control map to track request timestamps per IP
var requestTimestamps = make(map[string]time.Time)

func runServer() {
	// Ensure setup is complete before starting the server
	ensureSetup()

	// Initialize the router
	r := setupRouter()

	// Start the server
	startServer(r)
}

func startServer(r *mux.Router) {
	fmt.Println("Starting server on port 8085")
	server := &http.Server{
		Addr:    ":8085", // Updated to use HTTP only on port 8085
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server failed: %s\n", err)
		os.Exit(1)
	}
}

func withFloodControl(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if lastRequest, found := requestTimestamps[ip]; found {
			if time.Since(lastRequest) < 5*time.Second {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`<!DOCTYPE html><html><head>
					<link rel="stylesheet" href="/assets/css/bootstrap.min.css">
					<script src="/assets/js/bootstrap.bundle.min.js"></script>
					</head><body>
					<div class="container mt-5">
					<div class="alert alert-danger" role="alert">
					Too many requests. Please wait 5 seconds before trying again.
					</div></div></body></html>`))
				return
			}
		}
		requestTimestamps[ip] = time.Now()
		handler(w, r)
	}
}
