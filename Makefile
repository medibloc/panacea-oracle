export GO111MODULE = on

GO ?= ego-go

build_tags := $(strip $(BUILD_TAGS))
BUILD_FLAGS := -tags "$(build_tags)"

OUT_DIR = ./build

MODULE=github.com/medibloc/panacea-oracle

PROTO_DIR=proto
PROTOBUF_DIR=pb
PROTO_OUT_DIR=./

.PHONY: all build test sign-prod clean proto-gen

all: build test

build: go.sum
	$(GO) build -mod=readonly $(BUILD_FLAGS) -o $(OUT_DIR)/oracled ./cmd/oracled

test:
	docker pull ghcr.io/medibloc/panacea-core:main # since we use the tag 'main'
	$(GO) test -v ./...

lint:
	GO111MODULE=off go get github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run --timeout 5m0s --allow-parallel-runners
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

# Prepare ./scripts/private.pem that you want to use. If not, this command will generate a new one.
sign-prod:
	ego sign ./scripts/enclave-prod.json


clean:
	$(GO) clean
	rm -rf $(OUT_DIR)

proto-gen:
	rm -rfv $(PROTO_OUT_DIR)$(PROTOBUF_DIR)
	cd $(PROTO_DIR) && buf mod update && buf build && buf generate; cd -
