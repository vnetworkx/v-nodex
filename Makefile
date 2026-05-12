GO ?= go
CARGO ?= cargo

.PHONY: all fmt test build clean

all: test build

fmt:
	gofmt -w ./internal ./services ./configs ./cmd 2>/dev/null || true
	cd rust/kernel && $(CARGO) fmt --all

test:
	$(GO) test ./...
	cd rust/kernel && $(CARGO) test

build:
	$(GO) build ./services/node/cmd/vnodex-node
	$(GO) build ./services/node/cmd/vnodex-cli
	$(GO) build ./services/api/cmd/vnodex-api
	cd rust/kernel && $(CARGO) build

clean:
	rm -rf data
	cd rust/kernel && $(CARGO) clean
