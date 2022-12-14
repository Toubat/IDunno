IDUNNO_OUT := "./bin/idunno"
SDFS_OUT := "./bin/sdfs"
DNS_OUT := "./bin/dns"
BACKEND_OUT := "./bin/backend"
API_OUT := "./api/api.pb.go"
API_PB_OUT := "./api/api_grpc.pb.go"
PKG := "."
IDUNNO_PKG_BUILD := "${PKG}/idunno"
SDFS_PKG_BUILD := "${PKG}/sdfs"
DNS_PKG_BUILD := "${PKG}/dns"
BACKEND_PKG_BUILD := "${PKG}/backend"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all api idunno dns py backend

all: api idunno dns py backend

no-api: idunno dns py backend

api/api.pb.go: api/api.proto
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/api.proto

api: api/api.pb.go

py:
	@python -m grpc_tools.protoc -I./api --python_out=./inference --pyi_out=./inference --grpc_python_out=./inference ./api/api.proto

api_test:
	@protoc --go_out=. --go_opt=

dep: ## Get the dependencies
	@go get -v -d ./...

sdfs: dep api ## Build the binary file for sdfs
	@go build -v -o $(SDFS_OUT) $(SDFS_PKG_BUILD)

dns: dep api ## Build the binary file for dns
	@go build -v -o $(DNS_OUT) $(DNS_PKG_BUILD)

idunno: dep api
	@go build -v -o $(IDUNNO_OUT) $(IDUNNO_PKG_BUILD)

backend: dep api
	@go build -v -o $(BACKEND_OUT) $(BACKEND_PKG_BUILD)

clean: ## Remove previous builds
	@rm $(API_OUT) $(IDUNNO_OUT) $(DNS_OUT)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
