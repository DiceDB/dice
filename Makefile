VERSION := $(shell cat VERSION)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: build test build-docker run test-one

build:
	@echo "Building for $(GOOS)/$(GOARCH)"
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w -X config.DiceDBVersion=$(VERSION)" -o ./dicedb

build-debug:
	@echo "Building for $(GOOS)/$(GOARCH)"
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -gcflags="all=-N -l" -o ./dicedb

build-docker:
	docker build --tag dicedb/dicedb:latest --tag dicedb/dicedb:$(VERSION) .

##@ Testing

# Changing the parallel package count to 1 due to a possible race condition which causes the tests to get stuck.
# TODO: Fix the tests to run in parallel, and remove the -p=1 flag.
test: ## run the integration tests
	go clean -testcache
	CGO_ENABLED=1 go test -race -count=1 -p=1 ./tests/...

test-docs:
	go run ./scripts/test-docs/main.go

test-one: ## run a single integration test function by name (e.g. make test-one TEST_FUNC=TestSetGet)
	go clean -testcache
	CGO_ENABLED=1 go test -v -race -count=1 --run $(TEST_FUNC) ./tests/...

unittest: ## run the unit tests
	CGO_ENABLED=1 go test -race -count=1 ./internal/...

unittest-one: ## run a single unit test function by name (e.g. make unittest-one TEST_FUNC=TestSetGet)
	CGO_ENABLED=1 go test -race -count=1 --run $(TEST_FUNC) ./internal/...

##@ Development
run: ## run dicedb with the default configuration
	go run main.go --engine ironhawk --log-level debug --enable-wal

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
	go clean -cache -modcache
	rm -f dicedb
	@echo "Clean complete."

release: ## build and push the Docker image to Docker Hub with the latest tag and the version tag
	git tag $(VERSION)
	git push origin --tags
	$(MAKE) build-docker
	docker push dicedb/dicedb:$(VERSION)
	docker push dicedb/dicedb:latest

push-binary-remote:
	$(MAKE) build
	scp -i ${SSH_PEM_PATH} ./dicedb ubuntu@${REMOTE_HOST}:.

add-license-notice:
	./add_license_notice.sh

kill:
	@lsof -t -i :7379 | xargs -r kill -9

git-repair:
	find .git/objects/ -type f -empty | xargs rm
	git fetch -p
	git fsck --full

generate-docs:
	go run ./scripts/generate-docs/main.go
