package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

func main() {
	const filepathRoot = "."
	const port = "8080"
	apiCfg := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "%s\n", http.StatusText(http.StatusOK)); err != nil {
		log.Printf("handler readiness encountered an error: %v", err)
	}
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	const formatString = `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`

	body := fmt.Sprintf(formatString, cfg.fileserverHits.Load())
	if _, err := w.Write([]byte(body)); err != nil {
		log.Printf("handlerMetrics: write failed: %v", err)
	}
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("hits reset")); err != nil {
		log.Printf("handler reset encountered an error: %v", err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

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
		if err := respondWithError(w, 400, "chirp is too long"); err != nil {
			log.Printf("failed to error")
		}
		return
	}

	responseStruct := struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}{
		Valid:       true,
		CleanedBody: cleanedChirpString,
	}

	if err := respondWithJSON(w, 200, responseStruct); err != nil {
		log.Printf("error was encountered responding: %v\n", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}
