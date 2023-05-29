# Transparency Information Pod

This repository contains a small HTTP server called "tripod" or "transparency information pod". Upon being called, the transparency information pod gathers transparency information from each pod running in the default namespace. The returned JSON object is a snapshot of the system's current state of completeness regarding disclosing transparency requirements defined in the [GDPR](https://gdpr.eu/).

tripod is part of my Computer Science MSc. thesis at TU Berlin.

## Prerequisites

To be able to test tripod's functionalities, it needs a Kubernetes cluster. Tripod is developed for clusters running Kubernetes 1.24.0. I make no claims for Kubernetes versions above and below.

For building the image, [Go](https://golang.org) is required.

For now, I have been using self-signed certificates and cert-manager.
To issue and sign certificates, [cert-manager](https://cert-manager.io/) must be deployed to the cluster before tripod. To deploy cert-manager, these steps can be followed:
```
$ kubectl create namespace cert-manager # cert-manager is the default namespace
$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```
To verify the cert-manager api, run ```$cmctl check api```.

## Deploying tripod

### with ```kubectl````

tbd

### with Helm

A deployment chart can be found in isaschm/deployment-charts.

## Verfiying tripod runs

If tripod is deployed in its own namespace, its status can be verified with:
```
$ kubectl -n tripod get pods
NAME                     READY   STATUS    RESTARTS   AGE
tripod-dd66ddf4f-sgp4k   1/1     Running   0          31m
```

## Running tripod

For development, I used the [OpenTelemetry demo architecture](https://opentelemetry.io/docs/demo/kubernetes-deployment/) which has instructions on how to deploy using Helm.

Pods running in the "default" namespace don't need any prior tagging for tripod to return basic information.

As long as I have not provided a frontend (also tbd), a port-forward and curl returns the transparency information object.
Per default the server runs on port 80801. Hence, ```curl -k https://localhost:8081/map``` will do the trick.
For the demo architecture, this object might look like this:
```
[
    {
        "name":"my-otel-demo-accountingservice-d4cffbf49-4xkww",
        "dataCategories":"unspecified",
        "ttl":"0",
        "nodeLocation":"europe-north1"
    },
    {
        "name":"my-otel-demo-adservice-6fb45d559f-zbfnz",
        "dataCategories":"unspecified",
        "ttl":"0",
        "nodeLocation":"europe-north1"
    },
    ...
]
```