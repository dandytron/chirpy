package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dandytron/chirpy.git/internal/auth"
	"github.com/dandytron/chirpy.git/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	log.Printf("Validated JWT, user ID: %v", userID)

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		// send response "error": "Something went wrong"
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		// send response "error": "Something went wrong"
		return
	}

	// Check to make sure the Chirp isn't too long.
	const maxChirpLength = 140
	is_valid := isChirpTooLong(params.Body, maxChirpLength)
	if !is_valid {
		err = errors.New("this chirp is too long")
		respondWithError(w, http.StatusBadRequest, "Something went wrong:", err)
		return
	}

	// run chirp through the scrubber
	scrubbed_chirp := chirpScrubber(params.Body)
	newChirp, err := cfg.databaseQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   scrubbed_chirp,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp:", err)
		return
	}

	chirp := Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	}

	// send JSON response with scrubbed chirp, status 200
	respondWithJSON(w, http.StatusCreated, chirp)
}

// Helper function, checks to see if a 'chirp' is too long
func isChirpTooLong(chirp string, maxLength int) bool {
	return len(chirp) <= maxLength
}

// Assuming the length validation passed, replace any of the following words in the Chirp with the static 4-character string ****
func chirpScrubber(chirp string) string {
	split_chirp := strings.Split(chirp, " ")
	for i, word := range split_chirp {
		loweredWord := strings.ToLower(word)
		if _, ok := profaneWords[loweredWord]; ok {
			split_chirp[i] = "****"
		}
	}
	return strings.Join(split_chirp, " ")
}
