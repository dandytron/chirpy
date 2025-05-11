package main

import (
	"net/http"

	"github.com/google/uuid"
)

// Handler to retrieve all chirps in the database

func (cfg *apiConfig) retrieveAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.databaseQueries.RetrieveAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps: ", err)
		return
	}
	var chirpsSlice []Chirp
	for _, chirp := range chirps {
		retrievedChirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		chirpsSlice = append(chirpsSlice, retrievedChirp)
	}
	respondWithJSON(w, http.StatusOK, chirpsSlice)

}

// Handler to retrieve a single chirp in the database

func (cfg *apiConfig) retrieveSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not parse chirp ID:", err)
		return
	}

	chirp, err := cfg.databaseQueries.RetrieveSingleChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not retrieve chirp: ", err)
		return
	}

	retrievedChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, http.StatusOK, retrievedChirp)

}
