package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func clientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("create cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %v", err)
	}

	return clientset, nil
}

// Returns list of pods in "default" namespace
func getPods() (*v1.PodList, *kubernetes.Clientset, error) {
	clientset, err := clientSet()
	if err != nil {
		return nil, nil, fmt.Errorf("create clientset: %v", err)
	}

	podList, err := clientset.CoreV1().Pods("default").List(context.TODO(), meta.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("list pods: %v", err)
	}
	return podList, clientset, nil
}

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
