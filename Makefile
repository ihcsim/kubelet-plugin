SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64

build: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/device-plugin main.go

lint: tidy
	golangci-lint run .

tidy:
	go mod tidy

devices:
	mknod -m 666 /dev/fifo0 p
	mknod -m 666 /dev/fifo1 p
	mknod -m 666 /dev/fifo2 p

purge:
	rm /dev/fifo0 /dev/fifo1 /dev/fifo2
