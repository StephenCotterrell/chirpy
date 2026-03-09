package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/StephenCotterrell/chirpy/internal/auth"
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
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("error decoding params: %v", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	UserID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
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
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedChirpString,
		UserID: UserID,
	})
	if err != nil {
		respondWithError(w, 400, "failed to create chirp", err)
		return
	}

	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
