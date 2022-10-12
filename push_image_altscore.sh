#!/bin/bash
# This script is used to push the image to the docker hub
go build -o dapperdox
docker build -t dapperdox .
docker tag dapperdox:latest us-central1-docker.pkg.dev/credits-flow-staging/altscore-images/dapperdox:latest
docker push us-central1-docker.pkg.dev/credits-flow-staging/altscore-images/dapperdox:latest

