package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/StephenCotterrell/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUserUpgrade(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}
	authAPIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not parse auth header", err)
		return
	}

	if !cfg.validateAPIKey(authAPIKey) {
		respondWithError(w, http.StatusUnauthorized, "invalid api key", errors.New("invalid api key"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	if err = cfg.db.UserUpgradeChirpyRed(r.Context(), params.Data.UserID); err != nil {
		respondWithError(w, http.StatusNotFound, "user not found", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) validateAPIKey(apiKey string) bool {
	return apiKey == cfg.polkaAPIKey
}
