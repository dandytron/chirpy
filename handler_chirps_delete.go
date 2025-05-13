package main

import (
	"net/http"
	"strings"

	"github.com/dandytron/chirpy.git/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {

	// grab the chirp ID from the header
	// first, decode the request body
	chirpIDStr := r.PathValue("chirpID") // Or however your router provides path parameters
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", nil)
		return
	}

	// check for An access token in the header
	// If the access token is malformed or missing, respond with a 401 status code.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "Access token not found or malformed", nil)
		return
	}

	// Extract the actual token from the "Bearer " prefix
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	// Validate the JWT, grab the userID
	userID, err := auth.ValidateJWT(token, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not validate token", err)
		return
	}

	// Only allow the deletion of a chirp if the user is the author of the chirp.
	// First, retrieve chirp
	chirpToDelete, err := cfg.databaseQueries.RetrieveSingleChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp to delete not found", err)
		return
	}

	// Check if the user ID from the token matches the author ID of the chirp - if not, return 403
	if userID != chirpToDelete.UserID {
		respondWithError(w, http.StatusForbidden, "User mismatch, unauthorized to delete", nil)
		return
	}

	//If cannot be deleted, return a 500 (Internal Server Error) status code.
	err = cfg.databaseQueries.DeleteChirps(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete chirp", nil)
		return
	}
	//If the chirp is deleted successfully, return a 204 status code.

	w.WriteHeader(http.StatusNoContent)
}
