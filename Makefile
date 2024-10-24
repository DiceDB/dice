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

.DEFAULT_GOAL := build

# Help command
help:
	@echo "Available commands:"
	@echo "  build -               Build the project"
	@echo "  format -              Format the code"
	@echo "  run -                 Run the project"
	@echo "  test -                Run the integration tests"
	@echo "  test-one -            Run a specific integration tests"
	@echo "  unittest -            Run unit tests"
	@echo "  unittest-one -        Run a specific unit test"
	@echo "  build-docker -        Build docker image"
	@echo "  push-docker -         Push docker image"
	@echo "  lint -                Run linter"
	@echo "  run_benchmark -       Run benchmark"
	@echo "  run_benchmark-small - Run small benchmark"
	@echo "  run_benchmark-large - Run large benchmark"
	@echo "  clean -               Remove build artifacts"

build:
	@echo "Building for $(GOOS)/$(GOARCH)"
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./dicedb

format:
	go fmt ./...

run:
	go run main.go

# Changing the parallel package count to 1 due to a possible race condition which causes the tests to get stuck.
# TODO: Fix the tests to run in parallel, and remove the -p=1 flag.
test:
	go test -v -race -count=1 -p=1 ./integration_tests/...

test-one:
	go test -v -race -count=1 --run $(TEST_FUNC) ./integration_tests/...

unittest:
	go test -race -count=1 ./internal/...

unittest-one:
	go test -v -race -count=1 --run $(TEST_FUNC) ./internal/...

release:
	git tag $(VERSION)
	git push origin --tags
	docker build --tag dicedb/dicedb:latest --tag dicedb/dicedb:$(VERSION) .
	docker push dicedb/dicedb:$(VERSION)
	docker push dicedb/dicedb:latest

GOLANGCI_LINT_VERSION := 1.60.1

lint: check-golangci-lint
	golangci-lint run ./...

check-golangci-lint:
	@if ! command -v golangci-lint > /dev/null || ! golangci-lint version | grep -q "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Required golangci-lint version $(GOLANGCI_LINT_VERSION) not found."; \
		echo "Please install golangci-lint version $(GOLANGCI_LINT_VERSION) with the following command:"; \
		echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.1"; \
		exit 1; \
	fi

run_benchmark:
	@echo "Running memtier benchmark..."
	memtier_benchmark \
		--threads=$(THREADS) \
		--data-size=$(DATA_SIZE) \
		--key-pattern=$(KEY_PATTERN) \
		--clients=$(CLIENTS) \
		--requests=$(REQUESTS) \
		--port=$(PORT)
	@echo "Benchmark complete."

run-benchmark-small:
	$(MAKE) run_benchmark THREADS=2 DATA_SIZE=512 CLIENTS=20 REQUESTS=5000

run-benchmark-large:
	$(MAKE) run_benchmark THREADS=8 DATA_SIZE=4096 CLIENTS=100 REQUESTS=50000

clean:
	@echo "Cleaning build artifacts..."
	rm -f dicedb
	@echo "Clean complete."
