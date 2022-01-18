#!/bin/bash
mkdir app 
cd app
GOARCH=amd64 GOOS=linux go build ../../cmd/video_conferencing/server.go 
cd ..
tar czfv slug.tgz ./app