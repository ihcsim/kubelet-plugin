SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64

build: clean kvm

clean:
	rm -rf bin dist cdi.tar.gz

.PHONY: kvm
kvm: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/kvm-plugin cmd/kvm/main.go

lint: tidy
	golangci-lint run ./...

tidy:
	go mod tidy

test: tidy
	go test -v -race -cover ./...

run.kvm: kvm
	sudo ./bin/kvm-plugin

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
	sudo cp ./cdi/github.com.ihcsim.kvm.yaml /etc/cdi/github.com.ihcsim.kvm.yaml

.PHONY: deploy
deploy:
	mkdir -p kubelet/run/{pods,logs}
	cp yaml/busybox-*.yaml kubelet/run/pods

undeploy:
	rm kubelet/run/pods/*.yaml

cdi.tar.gz:
	tar -czvf cdi.tar.gz cdi
