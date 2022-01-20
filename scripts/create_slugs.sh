#!/bin/bash
mkdir app 
cd app
GOARCH=amd64 GOOS=linux go build ../cmd/video_conferencing/server.go 
cd ..
cp -r {config.json,migrations} ./app  
tar czfv slug.tgz ./app
