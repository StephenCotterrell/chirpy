package main

import (
	"log"
	"net/http"
	"os"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		respondWithError(w, 403, "forbidden", nil)
	}
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("hits reset")); err != nil {
		log.Printf("handler reset encountered an error: %v", err)
	}
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		log.Printf("there was an error reseting the users table: %v", err)
	}
}
