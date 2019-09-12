#!/bin/bash
while true; do

# for less verbose outout
export GIN_MODE=test

inotifywait -e modify,create,delete -r ./ && \
	clear
	go fmt ./... \
		&& go build -o build/twowaykv \
		&& go test  -count=1 ./... 
done
