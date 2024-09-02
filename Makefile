SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64

# see /proc/devices for the major number of /dev/random
# on your machine
MAJOR_VERSION_CRAND ?= 1
MINOR_VERSION_CRAND ?= 8

build: crand

crand: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/dp-crand cmd/crand/main.go

lint: tidy
	golangci-lint run ./...

tidy:
	go mod tidy

test: tidy
	go test -v -race -cover ./...

run: build
	sudo ./bin/device-plugin-crand

.PHONY: kubelet
kubelet:
	if [ ! -e kubelet/kubelet-v1.31.0 ]; then \
		tar -xzvf kubelet/kubelet-v1.31.0.tar.gz -C kubelet ;\
	fi ;\
	sudo kubelet/kubelet-v1.31.0 \
		--config=kubelet/kubelet.yaml \
		--hostname-override localhost \
		--v=4 2>&1 | tee kubelet/kubelet.log

.PHONY: cdi
cdi:
	sudo mkdir -p /etc/cdi
	sudo cp cdi/pflex.yaml /etc/cdi/pflex.yaml
	sudo rm -rf /dev/pflex[0-3]
	sudo mknod -m 666 /dev/pflex0 c 1 8
	sudo mknod -m 666 /dev/pflex1 c 1 8
	sudo mknod -m 666 /dev/pflex2 c 1 8

.PHONY: deploy
deploy:
	mkdir -p kubelet/run/{pods,logs}
	cp yaml/busybox.yaml kubelet/run/pods
