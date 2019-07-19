#!/bin/bash
while true; do

inotifywait -e modify,create,delete -r ./ && \
	clear
	go fmt ./... \
		&& go build -o build/twowaykv \
		&& go test ./... -v
done
