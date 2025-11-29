#!/bin/bash

mkdir -p test/bin/amd64/win
mkdir -p test/bin/amd64/lin
mkdir -p test/bin/amd64/mac
mkdir -p test/bin/arm64/win
mkdir -p test/bin/arm64/lin
mkdir -p test/bin/arm64/mac

env GOOS=linux GOARCH=amd64 go build -o test/bin/amd64/lin/pinshare .
env GOOS=windows GOARCH=amd64 go build -o test/bin/amd64/win/pinshare.exe .
env GOOS=darwin GOARCH=amd64 go build -o test/bin/amd64/mac/pinshare .
env GOOS=linux GOARCH=arm64 go build -o test/bin/arm64/lin/pinshare .
env GOOS=windows GOARCH=arm64 go build -o test/bin/arm64/win/pinshare.exe .
env GOOS=darwin GOARCH=arm64 go build -o test/bin/arm64/mac/pinshare .

