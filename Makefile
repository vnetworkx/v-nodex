APP_NODE := ./services/node/cmd/vnodex-node
APP_API := ./services/api/cmd/vnodex-api
APP_CLI := ./services/node/cmd/vnodex-cli

.PHONY: test build run node api cli fmt tidy zip

test:
	go test ./...

build:
	go build $(APP_NODE)
	go build $(APP_API)
	go build $(APP_CLI)

run:
	go run $(APP_NODE) --config configs/node.yaml

node:
	go run $(APP_NODE) --config configs/node.yaml

api:
	go run $(APP_API) --config configs/node.yaml

cli:
	go run $(APP_CLI) --config configs/node.yaml

fmt:
	gofmt -w ./

tidy:
	go mod tidy

zip:
	@echo "Use the prepared archive in /mnt/data after build."
