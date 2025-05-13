package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dandytron/chirpy.git/internal/auth"
	"github.com/dandytron/chirpy.git/internal/database"
)

func (cfg *apiConfig) updateCredentials(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string
	}
	type response struct {
		User
	}

	// check for An access token in the header
	// If the access token is malformed or missing, respond with a 401 status code.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "Access token not found or malformed", nil)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not validate token", err)
		return
	}

	// check for a new password and email in the request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode http request", err)
		return
	}
	if params.Email == "" || !strings.Contains(params.Email, "@") {
		respondWithError(w, http.StatusInternalServerError, "Email either not provided or invalid", nil)
		return
	}
	if params.Password == "" {
		respondWithError(w, http.StatusInternalServerError, "No password provided", nil)
		return
	}

	// Hash the password
	hashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password: %v", err)
		return
	}

	// update the hashed password and the email for the authenticated user in the database

	updatedUser, err := cfg.databaseQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	// Respond with a 200 if everything is successful and the newly
	// updated User resource (omitting the password of course).

	respondWithJSON(w, http.StatusOK, response{
		User{
			ID:        updatedUser.ID,
			CreatedAt: updatedUser.CreatedAt,
			UpdatedAt: updatedUser.UpdatedAt,
			Email:     updatedUser.Email,
		},
	})
}
