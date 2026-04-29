.PHONY: build build-cli build-agent build-all test test-atlassian-integration lint clean

BIN_DIR := .bin

build:
	go build -o $(BIN_DIR)/jiraforge ./cmd/jiraforge

build-cli:
	go build -o $(BIN_DIR)/jiraforge-cli ./cmd/jiraforge-cli

build-agent:
	go build -o $(BIN_DIR)/jiraforge-agent ./cmd/jiraforge-agent

build-all: build build-cli build-agent

test:
	go test ./...

test-atlassian-integration:
	go test -count=1 -tags=integration ./internal/adapters/atlassian -v

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)
