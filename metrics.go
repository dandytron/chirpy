package main

import (
	"fmt"
	"net/http"
)

// Middleware wrapper that helps track the state of server hits currently;
// returns a handler that increments the file server hits and calls the next handler
func (cfg *apiConfig) middlewareMetricsIncrementer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// apiConfig function that gets the atomic int32 value, converts it to a string, and sets "Hits: x"
// where x is the number of hits to the header of the response
func (cfg *apiConfig) numRequestHandler(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits.Load()
	hitsStr := fmt.Sprintf("Hits: %d", hits)
	w.Write([]byte(hitsStr))
}

// Function for serving simple HTML on an admin get request
func (cfg *apiConfig) adminMetricsHandler(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>`, hits)))
}
