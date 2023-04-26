package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
)

const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

// mapFuncHandler returns a http.Handler serving the map route by calling podTransparencyInformation
func mapFuncHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, r *http.Request) {
		podList, client, err := getPods()
		if err != nil {
			log.Fatalf("get pod list: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		pods, err := parseTransparencyInformation(podList, client)
		if err != nil {
			log.Fatalf("parse data categories: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		json.NewEncoder(writer).Encode(pods)
	}
}

func main() {
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/map", mapFuncHandler())
	server := &http.Server{
		Addr:    ":8081",
		Handler: withLogging(log.Default())(mux),
	}

	log.Printf("listening on %s", server.Addr)
	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}

func withLogging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			next.ServeHTTP(w, r)
		})
	}
}
