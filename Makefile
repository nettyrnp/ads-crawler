PROJECT_NAME := "ads-crawler"
PKG := "github.com/nettyrnp/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

all: build

lint: ## Lint the files
	@golint -set_exit_status ${PKG_LIST}

test: ## Run unittests
	@go test -tags=unit ${PKG_LIST}

test-integration: ## Run unittests
	@go test -tags=unit ${PKG_LIST}

race: ## Run data race detector
	@go test -race -short ${PKG_LIST}

build: ## Build the binary file
	GOOS=linux GOARCH=amd64 go build cmd/crawler.go

run: ## Run the app
	@go run cmd/crawler.go start -e .env

migrate:
	@go run cmd/crawler.go migrate -e .env

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
