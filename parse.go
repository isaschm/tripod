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
	Necessity      string      `json:"necessity"`
	AutoDecision   string      `json:"autoDecision"`
}

type DataCategory struct {
	Name       string `json:"name"`
	Purpose    string `json:"purpose,omitempty"`
	LegalBasis string `json:"legalBasis,omitempty"`
	Storage    string `json:"storage,omitempty"`
	Recipient  string `json:"recipient,omitempty"`
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

// parseTransparencyInformation gathers transparency information from every pod in podList
// It checks each pod for data categories, if non are given data categories are added
// as "unspecified". It also gets ttl and location from the node the pod runs on.
func parseTransparencyInformation(podList *v1.PodList, client *kubernetes.Clientset) ([]Tripod, error) {
	var pods []Tripod
	for _, pod := range podList.Items {
		// Get node via the pod node name
		node, err := client.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, meta.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("get node: %w", err)
		}

		labels := node.GetObjectMeta().GetLabels()

		// Parse pod annotations into a map to get easier access to tags
		var annotations map[string]string
		annotations = pod.Annotations

		countryIso := mapLocationKey(labels[locationKey])

		necessity, ok := annotations["necessity"]
		if !ok {
			necessity = unspecifiedTag
		}

		autoDecision, ok := annotations["autoDecision"]
		if !ok {
			autoDecision = unspecifiedTag
		}

		val, ok := annotations["dataCategories"]
		if !ok || val == unspecifiedTag {
			// If data categories are defined at all or tagged as "unspecified" by the
			// admission controller, return a tripod object with data categories as
			// "unspecified"
			pods = append(pods, Tripod{
				Name:           pod.Name,
				DataCategories: unspecifiedTag,
				NodeLocation:   countryIso,
				Ttl:            node.Annotations[ttlkey],
				Necessity:      necessity,
				AutoDecision:   autoDecision,
			})
		} else {
			// If data categories are defined, return the tripod object with data categories
			datacategories, err := parseDataCategories(val)
			if err != nil {
				return nil, fmt.Errorf("parsing data categories: %w", err)
			}
			pods = append(pods, Tripod{
				Name:           pod.Name,
				DataCategories: datacategories,
				NodeLocation:   countryIso,
				Ttl:            node.Annotations[ttlkey],
				Necessity:      necessity,
				AutoDecision:   autoDecision,
			})
		}
	}

	return pods, nil
}
