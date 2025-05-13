package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dandytron/chirpy.git/internal/auth"
	"github.com/dandytron/chirpy.git/internal/database"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode http request", nil)
		return
	}

	retrievedUser, err := cfg.databaseQueries.FindUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	err = auth.CheckPasswordHash(params.Password, retrievedUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	// Pass in time.Hour to represent "now" plus an hour's time.
	accessToken, err := auth.MakeJWT(
		retrievedUser.ID,
		cfg.jwtsecret,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	_, err = cfg.databaseQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    retrievedUser.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          retrievedUser.ID,
			CreatedAt:   retrievedUser.CreatedAt,
			UpdatedAt:   retrievedUser.UpdatedAt,
			Email:       retrievedUser.Email,
			IsChirpyRed: retrievedUser.IsChirpyRed.Bool,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}
