package main

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type Tripod struct {
	Name           string      `json:"name"`
	Purposes       string      `json:"purposes"`
	DataCategories interface{} `json:"dataCategories"`
}

type DataCategory struct {
	Name       string `json:"name"`
	Purpose    string `json:"purpose,omitempty"`
	LegalBasis string `json:"legalBasis,omitempty"`
	Storage    string `json:"stoagre,omitempty"`
}

const (
	unspecifiedTag = "unspecified"
)

func parseDataCategories(s string) ([]DataCategory, error) {
	categories := []DataCategory{}
	if err := json.Unmarshal([]byte(s), &categories); err != nil {
		return nil, fmt.Errorf("unmarshalling categories string: %w", err)
	}

	return categories, nil
}

func ParseTransparencyInformation(podList *v1.PodList) ([]Tripod, error) {
	var pods []Tripod
	for _, pod := range podList.Items {

		var annotations map[string]string
		annotations = pod.Annotations

		if annotations["dataCategories"] != unspecifiedTag {
			datacategories, err := parseDataCategories(annotations["dataCategories"])
			if err != nil {
				return nil, fmt.Errorf("parsing data categories: %w", err)
			}
			pods = append(pods, Tripod{
				Name:           pod.Name,
				Purposes:       annotations["purposes"],
				DataCategories: datacategories,
			})
		} else if annotations["dataCategories"] == unspecifiedTag {
			pods = append(pods, Tripod{
				Name:           pod.Name,
				Purposes:       annotations["purposes"],
				DataCategories: unspecifiedTag,
			})
		}
	}

	return pods, nil
}
