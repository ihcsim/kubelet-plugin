SHELL := /bin/bash

GOOS ?= linux
GOARCH ?= amd64

# see /proc/devices for the major number of /dev/random
# on your machine
MAJOR_VERSION_CRAND ?= 1
MINOR_VERSION_CRAND ?= 8

FLATCAR_DIR := ./flatcar

build: clean crand kvm

clean:
	rm -rf bin

crand: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/dp-crand cmd/crand/main.go

.PHONY: kvm
kvm: tidy
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/dp-kvm cmd/kvm/main.go

lint: tidy
	golangci-lint run ./...

tidy:
	go mod tidy

test: tidy
	go test -v -race -cover ./...

run.crand: crand
	sudo ./bin/dp-crand

run.kvm: kvm
	sudo ./bin/dp-kvm

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
	sed \
		-e 's/$${DEVICE_MAJOR_VERSION}/$(MAJOR_VERSION_CRAND)/g' \
		-e 's/$${DEVICE_MINOR_VERSION}/$(MINOR_VERSION_CRAND)/g' \
		./cdi/github.com.ihcsim.crand.yaml | sudo tee /etc/cdi/github.com.ihcsim.crand.yaml
	sudo rm -rf /dev/crand[0-3]
	sudo mknod -m 666 /dev/crand0 c $(MAJOR_VERSION_CRAND) $(MINOR_VERSION_CRAND)
	sudo mknod -m 666 /dev/crand1 c $(MAJOR_VERSION_CRAND) $(MINOR_VERSION_CRAND)
	sudo mknod -m 666 /dev/crand2 c $(MAJOR_VERSION_CRAND) $(MINOR_VERSION_CRAND)

.PHONY: deploy
deploy:
	mkdir -p kubelet/run/{pods,logs}
	cp yaml/busybox-*.yaml kubelet/run/pods

undeploy:
	rm kubelet/run/pods/*.yaml

$(FLATCAR_DIR):
	mkdir -p $(FLATCAR_DIR)
	wget https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img -P $(FLATCAR_DIR)
	qemu-img create -f qcow2 -F qcow2 -b flatcar_production_qemu_image.img $(FLATCAR_DIR)/flatcar-01.qcow2
	$(MAKE) butane2ign

butane2ign:
	sed -e "s/\$${SSH_PUB_KEY}/$$(cat ~/.ssh/id_ecdsa_gh.pub)/g" ./yaml/butane.yaml | docker run --rm -i quay.io/coreos/butane:release | jq . > ./flatcar/provision.ign

flatcar-start: $(FLATCAR_DIR)
	sudo virt-install \
    --connect qemu:///system \
    --osinfo generic \
    --import \
    --name flatcar-linux1 \
    --ram 1024 --vcpus 1 \
    --disk path=./flatcar/flatcar-linux1.qcow2,format=qcow2,bus=virtio,size=5 \
    --vnc --noautoconsole \
    --qemu-commandline='-fw_cfg name=opt/org.flatcar-linux/config,file=$(shell pwd)/flatcar/provision.ign'
