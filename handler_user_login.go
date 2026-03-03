package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/StephenCotterrell/chirpy/internal/auth"
)

func (cfg *apiConfig) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() //nolint:errcheck // nothing actionable on Close()

	type requestBody struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("failed to parse login request body")
		respondWithError(w, http.StatusInternalServerError, "failed to parse login request body", err)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "user not found", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not process password", err)
		return
	}

	if match {
		respondWithJSON(w, http.StatusOK, User{
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			ID:        user.ID,
		})
		return
	} else {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", nil)
		return
	}
}
