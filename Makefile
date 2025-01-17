PACKAGES=$(shell go list ./... | grep -v '/simulation')

VERSION := v0.2.2-dev
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=aura \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=aurad \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

all: install

install: go.sum
	@echo "--> Installing aurad"
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/aurad

build: go.sum
	@echo "--> Build aurad"
	@go build -mod=readonly $(BUILD_FLAGS) ./cmd/aurad

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

test:
	@go test -mod=readonly $(PACKAGES)
