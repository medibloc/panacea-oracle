export GO111MODULE = on

GO ?= ego-go

build_tags := $(strip $(BUILD_TAGS))
BUILD_FLAGS := -tags "$(build_tags)"

OUT_DIR = ./build

MODULE=github.com/medibloc/panacea-oracle

PROTO_DIR=proto
PROTO_OUT_DIR=./

.PHONY: all build test sign-prod clean

all: build test

build: go.sum
	$(GO) build -mod=readonly $(BUILD_FLAGS) -o $(OUT_DIR)/oracled ./cmd/oracled

test:
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
	protoc --proto_path=$(PROTO_DIR) \
		--proto_path=third_party/proto \
		--go_out=$(PROTO_OUT_DIR) \
			--go_opt=paths=import \
			--go_opt=module=$(MODULE) \
		--go-grpc_out=$(PROTO_OUT_DIR) \
			--go-grpc_opt=paths=import \
			--go-grpc_opt=module=$(MODULE) \
		--grpc-gateway_out=$(PROTO_OUT_DIR) \
			--grpc-gateway_opt logtostderr=true \
			--grpc-gateway_opt=paths=import \
			--grpc-gateway_opt=module=$(MODULE) \
		$(PROTO_DIR)/panacea_oracle/*/*/*.proto