SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64
KO_DOCKER_REPO ?= isim

build: clean kvm

clean:
	rm -rf bin dist cdi.tar.gz

.PHONY: kvm
kvm: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/kvm-device-plugin .

lint: tidy
	golangci-lint run ./...

tidy:
	go mod tidy

test: tidy
	go test -v -race -cover ./...

image:
	KO_DOCKER_REPO=$(KO_DOCKER_REPO) ko build -B .

image-debug:
	KO_DOCKER_REPO=$(KO_DOCKER_REPO) ko build -B --debug .

.PHONY: yaml
yaml: image
	KO_DOCKER_REPO=$(KO_DOCKER_REPO) ko resolve -B -f yaml/daemonset.yaml.tmpl > yaml/daemonset.yaml

run.kvm: kvm
	sudo ./bin/kvm-device-plugin

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
	cp yaml/kvm.yaml kubelet/run/pods

undeploy:
	rm kubelet/run/pods/kvm.yaml

cdi.tar.gz:
	tar -czvf cdi.tar.gz cdi
