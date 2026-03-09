package main

import (
	"net/http"

	"github.com/StephenCotterrell/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID != "" {
		cfg.handleSingleChirp(w, r, chirpID)
		return
	}

	var (
		dbChirps []database.Chirp
		err      error
	)

	authorIDStr := r.URL.Query().Get("author_id")
	sortStr := r.URL.Query().Get("sort")
	if authorIDStr != "" {
		authorID, parseErr := uuid.Parse(authorIDStr)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "invalid author_id", parseErr)
			return
		}
		dbChirps, err = cfg.db.GetChirpsByAuthor(r.Context(), authorID)
	} else {
		if sortStr == "" || sortStr == "asc" {
			dbChirps, err = cfg.db.GetChirps(r.Context())
		} else {
			dbChirps, err = cfg.db.GetChirpsDesc(r.Context())
		}
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to fetch chirps", err)
		return
	}

	chirps := make([]Chirp, 0, len(dbChirps))
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
}

func (cfg *apiConfig) handleSingleChirp(w http.ResponseWriter, r *http.Request, chirpID string) {
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
