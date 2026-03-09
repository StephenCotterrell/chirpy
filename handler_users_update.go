package main

import (
	"encoding/json"
	"net/http"

	"github.com/StephenCotterrell/chirpy/internal/auth"
	"github.com/StephenCotterrell/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't decode parameters", err)
		return
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to extract auth token from headers", err)
		return
	}

	userID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid JWT", err)
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create password hash", err)
		return
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: passwordHash,
		ID:             userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update user details", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			ID:          user.ID,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}
