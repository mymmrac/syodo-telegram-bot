# Adds $GOPATH/bit to $PATH
export PATH := $(PATH):$(shell go env GOPATH)/bin

help: ## Display this help message
	@echo "Usage:"
	@grep -E "^[a-zA-Z_-]+:.*? ## .+$$" $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}'

lint: ## Run golangci-lint
	golangci-lint run

lint-install: ## Install golangci-lint
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint-list: ## Run golangci-lint linters (print enabled & disabled linters)
	golangci-lint linters

test: ## Run tests
	go test ./...

build: ## Build binary
	GOOS=linux GOARCH=arm64 go build -o bin/syodo .

deploy-web:
	cd web && \
	yarn build && \
 	zip -r dist.zip dist && \
 	scp dist.zip ubuntu@mymm.gq:/home/ubuntu/syodo/ && \
 	ssh ubuntu@mymm.gq "cd syodo/ && rm -rf dist/ && unzip dist.zip"

deploy-bot: build
	ssh ubuntu@mymm.gq "sudo systemctl stop syodo-bot" && \
	scp text.toml ubuntu@mymm.gq:/home/ubuntu/syodo/ && \
	scp bin/syodo ubuntu@mymm.gq:/home/ubuntu/syodo/ && \
    ssh ubuntu@mymm.gq "sudo systemctl start syodo-bot"

.PHONY: help lint lint-install lint-list test build deploy-web deploy-bot
