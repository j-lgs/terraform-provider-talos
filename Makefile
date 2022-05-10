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

.PHONY: build fmt vet test clean install acctest local-dev-install vendor docs

all: build docs

vendor:
	@echo " -> Grabbing talos vendor code"
	mkdir -p vendor_talos
	git clone --depth=1 https://github.com/siderolabs/talos.git vendor_talos/talos
	mv talos vendor_talos/talos
	go mod vendor

fmt:
	@echo " -> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -path ./talos_vendor -prune -o -name '*.go' -print) | grep '^'

vet:
	@echo " -> vetting code"
	@go vet ./...

test:
	@echo " -> testing code"
	@go test -v ./talos

acc-test:
	@echo " -> acceptance testing code"

	@mkdir -p .tmp
	@wget -P .tmp -nc "https://github.com/siderolabs/talos/releases/download/v1.0.4/talos-amd64.iso"
	@pgrep http-server >/dev/null && echo "http-server still running on PID $$(pgrep http-server)" || http-server .tmp -p 8000 -s --no-dotfiles &

	TF_ACC=1 go test -v ./talos
	@kill "$$(pgrep http-server)"

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
