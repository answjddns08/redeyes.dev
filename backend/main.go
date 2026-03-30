package main

import (
	"backend/api"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := getEnv("PORT", ":5000")

	apiServer := api.New(log.Default())

	mux := http.NewServeMux()

	apiServer.Register(mux)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Starting server on %s", addr)

	log.Fatal(server.ListenAndServe())
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
