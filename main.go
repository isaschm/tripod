package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/fatih/structs"
)

const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

func calculateScore(pods []Tripod) map[string]int {
	score := make(map[string]int)

	score["numPods"] = len(pods)
	score["completePods"] = 0
	score["partialPods"] = 0

	for _, pod := range pods {
		m := structs.Map(pod)
		iter := reflect.ValueOf(m).MapRange()

	out:
		for iter.Next() {
			if iter.Value().Interface().(string) == unspecifiedTag {
				score["partialPods"] += 1
				break out
			}
			score["completePods"] += 1
		}
	}
	return score
}

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
			log.Fatalf("parse transparency information: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		score := calculateScore(pods)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(pods)
		json.NewEncoder(writer).Encode(score)
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
