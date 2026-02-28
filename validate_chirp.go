package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
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

	responseStruct := struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}{
		Valid:       true,
		CleanedBody: cleanedChirpString,
	}

	respondWithJSON(w, 200, responseStruct)
}
