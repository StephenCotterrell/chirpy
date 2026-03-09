package main

import (
	"net/http"

	"github.com/StephenCotterrell/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	chripIDPathVal := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chripIDPathVal)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to parse chirp UUID", err)
		return
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve auth token from headers", err)
		return
	}

	userID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "authentication failed, user unauthorized", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "only the author of the tweet can delete the tweet", err)
		return
	}

	if err = cfg.db.DeleteChirp(r.Context(), chirp.ID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to delete chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
