SHELL := /bin/bash

MAJOR := 0
MINOR := 0
MICRO := 0
NEXT_MICRO := 1
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(MICRO)

DATE                = $(shell date +'%d.%m.%Y')
TIME                = $(shell date +'%H:%M:%S')

KERNEL=$(shell if [ "$$(uname -s)" == "Linux" ]; then echo linux; fi)
ARCH=$(shell if [ "$$(uname -m)" == "x86_64" ]; then echo amd64; fi)

.PHONY: build fmt vet test clean install acctest local-dev-install

all: build

fmt:
	@echo " -> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -path ./talos_vendor -prune -o -name '*.go' -print) | grep '^'

vet:
	@echo " -> vetting code"
	@go vet ./...

test:
	@echo " -> testing code"
	@go test -v ./...

build: clean
	@echo " -> Building"
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -o bin/terraform-provider-talos
	@echo "Built terraform-provider-talos"

install: build
	cp bin/terraform-provider-talos $$GOPATH/bin/terraform-provider-talos

local-dev-install: build
	find examples -name '.terraform.lock.hcl' -delete
	@echo "$(CURRENT_VERSION_MICRO)"
	@echo "$(KERNEL)"
	@echo "$(ARCH)"
	mkdir -p ~/.terraform.d/plugins/localhost/jlgs/talos/$(MAJOR).$(MINOR).$(NEXT_MICRO)/$(KERNEL)_$(ARCH)/
	cp bin/terraform-provider-talos ~/.terraform.d/plugins/localhost/jlgs/talos/$(MAJOR).$(MINOR).$(NEXT_MICRO)/$(KERNEL)_$(ARCH)/terraform-provider-talos

#clean:
#	@git clean -f -d
