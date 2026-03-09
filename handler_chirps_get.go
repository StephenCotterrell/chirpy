package main

import (
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")

	if chirpID == "" {
		dbChirps, err := cfg.db.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to fetch all chirps", err)
			return
		}

		chirps := []Chirp{}
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.UpdatedAt,
				UserID:    dbChirp.UserID,
				Body:      dbChirp.Body,
			})
		}

		respondWithJSON(w, http.StatusOK, chirps)
		return
	}

	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to parse chirp UUID", err)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "no chirp found for chirp id", err)
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
	}

	respondWithJSON(w, http.StatusOK, chirp)
}
