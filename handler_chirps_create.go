package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/StephenCotterrell/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() //nolint:errcheck // nothing actionable on Close()

	type requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("error decoding params: %v", err)
		return
	}

	chirpString := params.Body
	chirpWords := strings.Split(chirpString, " ")
	cleanedChirpWords := []string{}
	bannedWords := map[string]struct{}{"kerfuffle": {}, "sharbert": {}, "fornax": {}}
	replacementString := "****"

	for _, chirpWord := range chirpWords {
		if _, banned := bannedWords[strings.ToLower(chirpWord)]; banned {
			cleanedChirpWords = append(cleanedChirpWords, replacementString)
		} else {
			cleanedChirpWords = append(cleanedChirpWords, chirpWord)
		}
	}

	cleanedChirpString := strings.Join(cleanedChirpWords, " ")

	if len(cleanedChirpString) > 140 {
		respondWithError(w, 400, "chirp is too long", nil)
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedChirpString,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, 400, "failed to create chirp", err)
	}

	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
