package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	locationMap = map[string]string{
		"asia-east1":              "TW",
		"asia-east2":              "HK",
		"asia-northeast1":         "JP",
		"asia-northeast2":         "JP",
		"asia-northeast3":         "KR",
		"asia-south1":             "IN",
		"asia-south2":             "IN",
		"asia-southeast1":         "SG",
		"asia-southeast2":         "ID",
		"australia-southeast1":    "AU",
		"australia-southeast2":    "AU",
		"europe-central2":         "PL",
		"europe-north1":           "FI",
		"europe-southwest1":       "ES",
		"europe-west1":            "BE",
		"europe-west12":           "IT",
		"europe-west2":            "GB",
		"europe-west3":            "DE",
		"europe-west4":            "NL",
		"europe-west6":            "CH",
		"europe-west8":            "IT",
		"europe-west9":            "FR",
		"me-central1":             "QA",
		"me-west1":                "IL",
		"northamerica-northeast1": "CA",
		"northamerica-northeast2": "CA",
		"southamerica-east1":      "BR",
		"southamerica-west1":      "CL",
		"us-central1":             "US",
		"us-east1":                "US",
		"us-east4":                "US",
		"us-east5":                "US",
		"us-south1":               "US",
		"us-west1":                "US",
		"us-west2":                "US",
		"us-west3":                "US",
		"us-west4":                "US",
	}
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

func mapLocationKey(location string) string {
	return locationMap[location]
}
