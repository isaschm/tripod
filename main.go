package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
)

const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

type Score struct {
	NumPods             int      `json:"numPods"`
	IncompletePodsCount int      `json:"partialPods"`
	IncompletePods      []string `json:"incompletePods,omitempty"`
}

func calculateScore(pods []Tripod) Score {
	score := Score{NumPods: len(pods), IncompletePodsCount: 0}
	incompletePods := []string{}

ScoreLoop:
	for _, pod := range pods {
		v := reflect.ValueOf(pod)

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.Type().String() == "string" && field.String() == unspecifiedTag {
				score.IncompletePodsCount += 1
				incompletePods = append(incompletePods, pod.Name)
				continue ScoreLoop
			}
		}
	}
	if len(incompletePods) > 0 {
		score.IncompletePods = incompletePods
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
