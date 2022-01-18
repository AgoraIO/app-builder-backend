#!/bin/bash

GOARCH=amd64 GOOS=linux go build -o appBuilderCore ../cmd/video_conferencing/server.go 
tar czfv appBuilderCore.tgz ./appBuilderCore