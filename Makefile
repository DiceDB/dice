.PHONY: build test build-docker run test-one

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./dice-server

format:
	go fmt ./...

run:
	go run main.go

test:
	go test -v -count=1 ./tests/

test-one:
	go test -v -count=1 --run $(TEST_FUNC) ./tests/...

unittest:
	go test -v -count=1 ./core/...

unittest-one:
	go test -v -count=1 --run $(TEST_FUNC) ./core/...

build-docker:
	docker build --tag dicedb/dice-server:latest --tag dicedb/dice-server:0.0.1 .

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
