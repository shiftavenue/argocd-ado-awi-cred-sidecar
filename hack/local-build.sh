#!/bin/bash
# the first two inputs are the version and destination registry
VERSION=$1
REGISTRY=$2

CGO_ENABLED=0 go build -o ./argocd-ado-awi-cred-sidecar cmd/main.go

docker build -t $REGISTRY/argocd-ado-awi-cred-sidecar:$VERSION .
docker push $REGISTRY/argocd-ado-awi-cred-sidecar:$VERSION