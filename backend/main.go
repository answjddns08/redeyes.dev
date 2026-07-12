package main

import (
	"log"
	"net/http"
	"os"

	"backend/api"
)

func main() {
	addr := getEnv("PORT", ":5000")

	apiServer := api.New(log.Default())

	mux := http.NewServeMux()

	apiServer.Register(mux)

	server := &http.Server{
		Addr:    addr,
		Handler: withCORS(withLogging(mux)),
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

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173":     {},
		"https://blog.redeyes.dev":  {},
		"http://192.168.35.11:5173": {},
		"app://obsidian.md":         {},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
