package main

import (
	"net/http"

	"github.com/StephenCotterrell/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse authorization header", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to revoke token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
