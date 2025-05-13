package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dandytron/chirpy.git/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) chirpyRedHandler(w http.ResponseWriter, r *http.Request) {

	// Define a struct to match the webhook payload
	type PolkaWebhook struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	// Check for an API key in the header. Format: Authorization: ApiKey THE_KEY_HERE

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "ApiKey ") {
		respondWithError(w, http.StatusUnauthorized, "API key not found or malformed", nil)
		return
	}
	webhookApiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch API key from JSON header", err)
		return
	}
	fmt.Println("Received API key:", webhookApiKey)
	fmt.Println("Expected API key:", cfg.polkaAPIKey)
	fmt.Println("Are they equal?", webhookApiKey == cfg.polkaAPIKey)
	if webhookApiKey != cfg.polkaAPIKey {
		respondWithError(w, http.StatusUnauthorized, "API key does not match our records", err)
		return
	}

	// Grab the webhook params from the request body
	decoder := json.NewDecoder(r.Body)
	webhook := PolkaWebhook{}
	err = decoder.Decode(&webhook)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// If the event is anything other than user.upgraded, the endpoint should immediately
	// respond with a 204 status code - we don't care about any other events.
	if webhook.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If the event is user.upgraded, then it should update the user in the database
	// and mark that they are a Chirpy Red member.

	uuidUserID, err := uuid.Parse(webhook.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "ID string could not be parsed into a UUID", err)
		return
	}

	err = cfg.databaseQueries.UpgradeToChirpyRed(r.Context(), uuidUserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
