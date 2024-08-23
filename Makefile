SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64

build: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/device-plugin main.go

lint: tidy
	golangci-lint run .

tidy:
	go mod tidy

run: build
	sudo ./bin/device-plugin

.PHONY: kubelet
kubelet:
	if [ ! -e kubelet/kubelet-v1.31.0 ]; then \
		tar -xzvf kubelet/kubelet-v1.31.0.tar.gz -C kubelet ;\
	fi ;\
	sudo kubelet/kubelet-v1.31.0 \
		--config=kubelet/kubelet.yaml \
		--hostname-override localhost \
		--feature-gates=ResourceHealthStatus=true \
		--v=4 2>&1 | tee kubelet/kubelet.log

.PHONY: cdi
cdi:
	sudo mkdir -p /etc/cdi
	sudo cp cdi/pflex.yaml /etc/cdi/pflex.yaml
	sudo rm -rf /dev/pflex*
	sudo mknod -m 666 /dev/pflex0 b 25 25
	sudo mknod -m 666 /dev/pflex1 b 25 25
	sudo mknod -m 666 /dev/pflex2 b 25 25
