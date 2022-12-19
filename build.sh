#!/usr/bin/env bash

# Set -E to stop if any command other than conditional command fails to execute
set -e

# Name of the docker image
IMAGE=$1;

# Build ID for the intermediate images
BUILD_ID=`uuidgen`

echo $BUILD_ID

if [[ -z "$1" ]]; then
#    echo -e "\nPlease call '$0 <image>' to deploy this image!\n"
#    exit 1
     IMAGE="agora/appbuilder"
fi

# type of the image
DOCKER_FILE="Dockerfile"


if [[ "dev" == "$2" ]]; then
    DOCKER_FILE="Dockerfile-Dev"
fi

#echo "${envs[*]}"

docker build \
    --build-arg BUILD_ID=$BUILD_ID \
    -t \
    $IMAGE \
    -f ${DOCKER_FILE} .

# Filter out and remove the intermediate build
docker image prune --force --filter label=stage=Stage1 --filter label=build=$BUILD_ID
