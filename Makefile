# Adapted from https://www.thapaliya.com/en/writings/well-documented-makefiles/

THREADS ?= 4 #number of threads
CLIENTS ?= 50 #number of clients per thread
REQUESTS ?= 10000 #number of requests per client
DATA_SIZE ?= 32 #Object data size
KEY_PATTERN ?= R:R #Set:Get pattern
RATIO ?= 1:10 #Set:Get ratio
PORT ?= 7379 #Port for dicedb

# Default OS and architecture
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

VERSION=$(shell bash -c 'grep -oP "DiceDBVersion string = \"\K[^\"]+" config/config.go')

.PHONY: build test build-docker run test-one

.DEFAULT_GOAL := help

##@ Helpers

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Building

build: ## generate the dicedb binary for the current OS and architecture
	@echo "Building for $(GOOS)/$(GOARCH)"
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./dicedb

build-debug: ## generate the dicedb binary for the current OS and architecture
	@echo "Building for $(GOOS)/$(GOARCH)"
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -gcflags="all=-N -l" -o ./dicedb

##@ Testing

# Changing the parallel package count to 1 due to a possible race condition which causes the tests to get stuck.
# TODO: Fix the tests to run in parallel, and remove the -p=1 flag.
test: ## run the integration tests
	go test -race -count=1 -p=1 ./integration_tests/...

test-one: ## run a single integration test function by name (e.g. make test-one TEST_FUNC=TestSetGet)
	go test -v -race -count=1 --run $(TEST_FUNC) ./integration_tests/...

unittest: ## run the unit tests
	go test -race -count=1 ./internal/...

unittest-one: ## run a single unit test function by name (e.g. make unittest-one TEST_FUNC=TestSetGet)
	go test -v -race -count=1 --run $(TEST_FUNC) ./internal/...

##@ Benchmarking

run_benchmark: ## run the memtier benchmark with the specified parameters
	@echo "Running memtier benchmark..."
	memtier_benchmark \
		--threads=$(THREADS) \
		--data-size=$(DATA_SIZE) \
		--key-pattern=$(KEY_PATTERN) \
		--clients=$(CLIENTS) \
		--requests=$(REQUESTS) \
		--port=$(PORT)
	@echo "Benchmark complete."

run-benchmark-small: ## run the memtier benchmark with small parameters
	$(MAKE) run_benchmark THREADS=2 DATA_SIZE=512 CLIENTS=20 REQUESTS=5000

run-benchmark-large: ## run the memtier benchmark with large parameters
	$(MAKE) run_benchmark THREADS=8 DATA_SIZE=4096 CLIENTS=100 REQUESTS=50000


##@ Development
run: ## run dicedb with the default configuration
	go run main.go

run-docker: ## run dicedb in a Docker container
	docker run -p 7379:7379 dicedb/dicedb:latest

format: ## format the code using go fmt
	go fmt ./...

GOLANGCI_LINT_VERSION := 1.60.1

lint:
	gofmt -w .
	golangci-lint run ./...

check-golangci-lint:
	@if ! command -v golangci-lint > /dev/null || ! golangci-lint version | grep -q "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Required golangci-lint version $(GOLANGCI_LINT_VERSION) not found."; \
		echo "Please install golangci-lint version $(GOLANGCI_LINT_VERSION) with the following command:"; \
		echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.1"; \
		exit 1; \
	fi

clean: ## clean the dicedb binary
	@echo "Cleaning build artifacts..."
	rm -f dicedb
	@echo "Clean complete."

##@ Deployment

release: ## build and push the Docker image to Docker Hub with the latest tag and the version tag
	git tag $(VERSION)
	git push origin --tags
	docker build --tag dicedb/dicedb:latest --tag dicedb/dicedb:$(VERSION) .
	docker push dicedb/dicedb:$(VERSION)
	docker push dicedb/dicedb:latest

push-binary-remote:
	$(MAKE) build
	scp -i ${SSH_PEM_PATH} ./dicedb ubuntu@${REMOTE_HOST}:.

add-license-notice:
	./add_license_notice.sh
