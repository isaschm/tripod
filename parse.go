package main

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Tripod struct {
	Name           string      `json:"name"`
	DataCategories interface{} `json:"dataCategories"`
	Ttl            string      `json:"ttl"`
	NodeLocation   string      `json:"nodeLocation"`
}

type DataCategory struct {
	Name       string `json:"name"`
	Purpose    string `json:"purpose,omitempty"`
	LegalBasis string `json:"legalBasis,omitempty"`
	Storage    string `json:"stoagre,omitempty"`
}

const (
	unspecifiedTag = "unspecified"
	locationKey    = "topology.kubernetes.io/region"
	ttlkey         = "node.alpha.kubernetes.io/ttl"
)

func parseDataCategories(s string) ([]DataCategory, error) {
	categories := []DataCategory{}
	if err := json.Unmarshal([]byte(s), &categories); err != nil {
		return nil, fmt.Errorf("unmarshalling categories string: %w", err)
	}

	return categories, nil
}

func parseTransparencyInformation(podList *v1.PodList, client *kubernetes.Clientset) ([]Tripod, error) {
	var pods []Tripod
	for _, pod := range podList.Items {
		node, err := client.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, meta.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("get node: %w", err)
		}

		labels := node.GetObjectMeta().GetLabels()

		var annotations map[string]string
		annotations = pod.Annotations

		val, ok := annotations["dataCategories"]
		if !ok || val == unspecifiedTag {
			pods = append(pods, Tripod{
				Name:           pod.Name,
				DataCategories: unspecifiedTag,
				NodeLocation:   labels[locationKey],
				Ttl:            node.Annotations[ttlkey],
			})
		} else {
			datacategories, err := parseDataCategories(val)
			if err != nil {
				return nil, fmt.Errorf("parsing data categories: %w", err)
			}
			pods = append(pods, Tripod{
				Name:           pod.Name,
				DataCategories: datacategories,
				NodeLocation:   labels[locationKey],
				Ttl:            node.Annotations[ttlkey],
			})
		}
	}

	return pods, nil
}
