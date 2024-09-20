THREADS ?= 4 #number of threads
CLIENTS ?= 50 #number of clients per thread
REQUESTS ?= 10000 #number of requests per client
DATA_SIZE ?= 32 #Object data size 
KEY_PATTERN ?= R:R #Set:Get pattern
RATIO ?= 1:10 #Set:Get ratio
PORT ?= 7379 #Port for dicedb

.PHONY: build test build-docker run test-one

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./dicedb

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

build-docker:
	docker build --tag dicedb/dicedb:latest --tag dicedb/dicedb:0.0.4 .

push-docker:
	docker push dicedb/dicedb:0.0.4

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
