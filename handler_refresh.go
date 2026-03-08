package main

import (
	"net/http"
	"time"

	"github.com/StephenCotterrell/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() //nolint:errcheck // nothing actionable on Close()

	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse authorization header", err)
	}

	dbRefreshRecord, err := cfg.db.CheckRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error checking refresh token record", err)
		return
	}

	if time.Now().After(dbRefreshRecord.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "token is expired", nil)
		return
	}

	if dbRefreshRecord.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "token is revoked", nil)
		return
	}

	token, err := auth.MakeJWT(dbRefreshRecord.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to make jwt", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}
