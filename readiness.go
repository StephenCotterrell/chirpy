package main

import (
	"fmt"
	"log"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "%s\n", http.StatusText(http.StatusOK)); err != nil {
		log.Printf("handler readiness encountered an error: %v", err)
	}
}
