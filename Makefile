SHELL := /bin/bash

MAJOR := 0
MINOR := 0
MICRO := 0
NEXT_MICRO := 10
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(MICRO)

DATE                = $(shell date +'%d.%m.%Y')
TIME                = $(shell date +'%H:%M:%S')

KERNEL=$(shell if [ "$$(uname -s)" == "Linux" ]; then echo linux; fi)
ARCH=$(shell if [ "$$(uname -m)" == "x86_64" ]; then echo amd64; fi)

.PHONY: build fmt vet check test clean install acctest local-dev-install docs

all: check test build docs

fmt:
	@echo " -> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -path ./talos_vendor -prune -o -name '*.go' -print) | grep '^'

vet:
	@echo " -> vetting code"
	@go vet ./...

test:
	@echo " -> testing code"
	@go test -v -race -vet=off ./...

check:
	@echo " -> checking code"
	@staticcheck ./...
	@golint $(go list ./... | grep -v /vendor\|vendor_talos/)

acctest:
	@echo " -> acceptance testing code"
	RESET_VM=1 TF_ACC=1 MACHINELOG_DIR=$$(pwd) TALOSCONF_DIR=$$(pwd) REGISTRY_CACHE=$$(pwd)/.registrycache go test -v ./talos
	tools/cleanAcceptance.sh

build:
	@echo " -> Building"
	goreleaser build --rm-dist --single-target --snapshot
	@echo "Built terraform-provider-talos"

docs:
	tfplugindocs

install: build
	cp dist/terraform-provider-talos_linux_amd64_v1/terraform-provider-talos_* $$GOPATH/bin/terraform-provider-talos

local-dev-install: build
	find examples -name '.terraform.lock.hcl' -delete
	@echo "$(CURRENT_VERSION_MICRO)"
	@echo "$(KERNEL)"
	@echo "$(ARCH)"
	mkdir -p ~/.terraform.d/plugins/localhost/j-lgs/talos/$(MAJOR).$(MINOR).$(NEXT_MICRO)/$(KERNEL)_$(ARCH)/
	cp dist/terraform-provider-talos_linux_amd64_v1/terraform-provider-talos_* ~/.terraform.d/plugins/localhost/j-lgs/talos/$(MAJOR).$(MINOR).$(NEXT_MICRO)/$(KERNEL)_$(ARCH)/terraform-provider-talos
