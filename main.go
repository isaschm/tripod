package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

// Returns list of pods in "default" namespace
func getPods() (*v1.PodList, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("create cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %v", err)
	}

	podList, err := clientset.CoreV1().Pods("default").List(context.TODO(), meta.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %v", err)
	}
	return podList, nil
}

func dashBoardHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, r *http.Request) {

		podList, err := getPods()
		if err != nil {
			log.Fatalf("get pod list: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		var names []string
		for _, pod := range podList.Items {
			names = append(names, pod.Name)
		}

		fmt.Fprint(writer, names)
	}
}

func main() {
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/map", dashBoardHandler())
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
