package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dandytron/chirpy.git/internal/auth"
	"github.com/dandytron/chirpy.git/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request
	type parameters struct {
		Email    string `json:"email"`
		Password string
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Println("createUserHandler: error decoding:", err)
		respondWithError(w, http.StatusInternalServerError, "Could not decode http request", err)
		return
	}

	// Hash password
	hashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password: %v", err)
		return
	}

	// Call SQLC to create user
	dbUser, err := cfg.databaseQueries.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		HashedPassword: hashedPW,
		Email:          params.Email,
	})
	if err != nil {
		log.Println("createUserHandler: error creating user:", err)
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}

	// Convert to response model, use respondWithJson function to send response
	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, http.StatusCreated, user)
}
